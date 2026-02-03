package workflow

import (
	"encoding/json"
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

type Document struct {
	Name       string              `json:"name" yaml:"name"`
	TimeoutSec int                 `json:"timeout_sec" yaml:"timeout_sec"`
	FailFast   bool                `json:"fail_fast" yaml:"fail_fast"`
	Nodes      []domain.WorkflowNode `json:"nodes" yaml:"nodes"`
	Edges      []domain.WorkflowEdge `json:"edges" yaml:"edges"`
}

func FromDocument(namespaceID string, doc Document) *domain.Workflow {
	now := time.Now().UTC()
	return &domain.Workflow{
		ID:          domain.NewID(),
		NamespaceID: namespaceID,
		Name:        doc.Name,
		Nodes:       doc.Nodes,
		Edges:       doc.Edges,
		TimeoutSec:  doc.TimeoutSec,
		FailFast:    doc.FailFast,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func ParseJSON(raw []byte) (Document, error) {
	var doc Document
	err := json.Unmarshal(raw, &doc)
	return doc, err
}

func NodeByID(nodes []domain.WorkflowNode, id string) (*domain.WorkflowNode, bool) {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i], true
		}
	}
	return nil, false
}

func Upstream(edges []domain.WorkflowEdge, nodeID string) []string {
	var out []string
	for _, e := range edges {
		if e.To == nodeID {
			out = append(out, e.From)
		}
	}
	return out
}

func Downstream(edges []domain.WorkflowEdge, nodeID string) []string {
	var out []string
	for _, e := range edges {
		if e.From == nodeID {
			out = append(out, e.To)
		}
	}
	return out
}
