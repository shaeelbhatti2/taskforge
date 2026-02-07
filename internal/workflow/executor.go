package workflow

import (
	"context"
	"sync"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

type NodeRunner func(ctx context.Context, node domain.WorkflowNode) error

type Executor struct {
	run NodeRunner
}

func NewExecutor(run NodeRunner) *Executor {
	return &Executor{run: run}
}

func (e *Executor) Run(ctx context.Context, wf *domain.Workflow) error {
	if err := Validate(wf); err != nil {
		return err
	}
	order, err := TopologicalSort(wf.Nodes, wf.Edges)
	if err != nil {
		return err
	}
	for _, nodeID := range order {
		node, ok := NodeByID(wf.Nodes, nodeID)
		if !ok {
			continue
		}
		if node.Type == domain.NodeFork {
			if err := e.runParallel(ctx, wf, nodeID); err != nil {
				if wf.FailFast {
					return err
				}
			}
			continue
		}
		if err := e.run(ctx, *node); err != nil && wf.FailFast {
			return err
		}
	}
	return nil
}

func (e *Executor) runParallel(ctx context.Context, wf *domain.Workflow, forkID string) error {
	children := Downstream(wf.Edges, forkID)
	var wg sync.WaitGroup
	errCh := make(chan error, len(children))
	for _, childID := range children {
		node, ok := NodeByID(wf.Nodes, childID)
		if !ok {
			continue
		}
		wg.Add(1)
		go func(n domain.WorkflowNode) {
			defer wg.Done()
			if err := e.run(ctx, n); err != nil {
				errCh <- err
			}
		}(*node)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func TopologicalSort(nodes []domain.WorkflowNode, edges []domain.WorkflowEdge) ([]string, error) {
	inDegree := map[string]int{}
	adj := map[string][]string{}
	for _, n := range nodes {
		inDegree[n.ID] = 0
		adj[n.ID] = []string{}
	}
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e.To)
		inDegree[e.To]++
	}
	queue := []string{}
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	order := []string{}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		order = append(order, cur)
		for _, next := range adj[cur] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	if len(order) != len(nodes) {
		return nil, domainErr("cycle detected during sort")
	}
	return order, nil
}

func domainErr(msg string) error {
	return &sortError{msg: msg}
}

type sortError struct{ msg string }

func (e *sortError) Error() string { return e.msg }
