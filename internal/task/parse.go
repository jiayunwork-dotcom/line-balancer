package task

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseCSV reads a precedence-constrained task set from a CSV reader.
// Expected columns: task_id, seconds, predecessors (semicolon-separated IDs).
// The first row is treated as a header if the first cell (lowercased, trimmed)
// matches "task", "id", or "工序".
func ParseCSV(r io.Reader) (*Graph, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true
	recs, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(recs) == 0 {
		return nil, fmt.Errorf("empty CSV")
	}

	// Detect and skip header row.
	start := 0
	if len(recs[0]) > 0 {
		h := strings.TrimSpace(strings.ToLower(strings.TrimPrefix(recs[0][0], "\ufeff")))
		if h == "task" || h == "id" || h == "工序" || h == "task_id" || h == "name" {
			start = 1
		}
	}

	g := NewGraph()
	type pendingEdge struct{ from, to string }
	var edges []pendingEdge

	for i := start; i < len(recs); i++ {
		rec := recs[i]
		if len(rec) < 2 {
			if len(rec) == 0 || strings.TrimSpace(rec[0]) == "" {
				continue
			}
			return nil, fmt.Errorf("line %d: expected at least 'id,seconds', got %q", i+1, strings.Join(rec, ","))
		}
		id := strings.TrimSpace(rec[0])
		if id == "" {
			continue
		}
		t, err := strconv.ParseFloat(strings.TrimSpace(rec[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid seconds %q", i+1, rec[1])
		}
		if err := g.Add(id, t); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		if len(rec) >= 3 {
			preds := strings.TrimSpace(rec[2])
			if preds != "" && preds != "-" {
				for _, p := range strings.Split(preds, ";") {
					p = strings.TrimSpace(p)
					if p != "" {
						edges = append(edges, pendingEdge{from: p, to: id})
					}
				}
			}
		}
	}

	for _, e := range edges {
		if err := g.Precede(e.from, e.to); err != nil {
			return nil, fmt.Errorf("precedence %q->%q: %w", e.from, e.to, err)
		}
	}
	if err := g.Validate(); err != nil {
		return nil, err
	}
	return g, nil
}

// FormatCSV writes the graph as a CSV with columns: task_id, seconds, predecessors.
func FormatCSV(g *Graph, w io.Writer) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"task_id", "seconds", "predecessors"}); err != nil {
		return err
	}
	for _, id := range g.IDs() {
		t := g.Get(id)
		preds := strings.Join(t.Predecessors, ";")
		if preds == "" {
			preds = "-"
		}
		rec := []string{id, strconv.FormatFloat(t.Time, 'f', -1, 64), preds}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
