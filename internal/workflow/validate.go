package workflow

import (
	"fmt"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

func Validate(wf *domain.Workflow) error {
	if wf == nil {
		return fmt.Errorf("workflow is nil")
	}
	if len(wf.Nodes) == 0 {
		return fmt.Errorf("workflow has no nodes")
	}
	ids := map[string]struct{}{}
	for _, n := range wf.Nodes {
		if n.ID == "" {
			return fmt.Errorf("node missing id")
		}
		if _, ok := ids[n.ID]; ok {
			return fmt.Errorf("duplicate node id %s", n.ID)
		}
		ids[n.ID] = struct{}{}
	}
	for _, e := range wf.Edges {
		if _, ok := ids[e.From]; !ok {
			return fmt.Errorf("edge from unknown node %s", e.From)
		}
		if _, ok := ids[e.To]; !ok {
			return fmt.Errorf("edge to unknown node %s", e.To)
		}
	}
	if err := detectCycle(wf.Nodes, wf.Edges); err != nil {
		return err
	}
	return nil
}

func detectCycle(nodes []domain.WorkflowNode, edges []domain.WorkflowEdge) error {
	adj := map[string][]string{}
	inDegree := map[string]int{}
	for _, n := range nodes {
		adj[n.ID] = []string{}
		inDegree[n.ID] = 0
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
	visited := 0
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		visited++
		for _, next := range adj[cur] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	if visited != len(nodes) {
		return fmt.Errorf("workflow contains a cycle")
	}
	return nil
}
