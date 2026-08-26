package task

import "fmt"

type Task struct {
	ID           string
	Time         float64
	Successors   []string
	Predecessors []string
}

type Graph struct {
	tasks map[string]*Task
	order []string
}

func NewGraph() *Graph {
	return &Graph{tasks: make(map[string]*Task)}
}

func (g *Graph) Add(id string, time float64) error {
	if _, ok := g.tasks[id]; ok {
		return fmt.Errorf("duplicate task %q", id)
	}
	if time < 0 {
		return fmt.Errorf("task %q: negative time %v", id, time)
	}
	g.tasks[id] = &Task{ID: id, Time: time}
	g.order = append(g.order, id)
	return nil
}

func (g *Graph) Precede(from, to string) error {
	f, ok := g.tasks[from]
	if !ok {
		return fmt.Errorf("unknown task %q", from)
	}
	t, ok := g.tasks[to]
	if !ok {
		return fmt.Errorf("unknown task %q", to)
	}
	f.Successors = append(f.Successors, to)
	t.Predecessors = append(t.Predecessors, from)
	return nil
}

func (g *Graph) Get(id string) *Task {
	return g.tasks[id]
}

func (g *Graph) IDs() []string {
	out := make([]string, len(g.order))
	copy(out, g.order)
	return out
}

func (g *Graph) Size() int { return len(g.tasks) }

func (g *Graph) Validate() error {
	for _, t := range g.tasks {
		for _, s := range t.Successors {
			if _, ok := g.tasks[s]; !ok {
				return fmt.Errorf("task %q references unknown successor %q", t.ID, s)
			}
		}
		for _, p := range t.Predecessors {
			if _, ok := g.tasks[p]; !ok {
				return fmt.Errorf("task %q references unknown predecessor %q", t.ID, p)
			}
		}
	}
	_, err := g.TopologicalOrder()
	return err
}

func (g *Graph) TopologicalOrder() ([]string, error) {
	inDeg := make(map[string]int, len(g.tasks))
	for _, t := range g.tasks {
		if _, ok := inDeg[t.ID]; !ok {
			inDeg[t.ID] = 0
		}
		for _, s := range t.Successors {
			inDeg[s]++
		}
	}
	var queue []string
	for _, id := range g.order {
		if inDeg[id] == 0 {
			queue = append(queue, id)
		}
	}
	var result []string
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		result = append(result, cur)
		for _, s := range g.tasks[cur].Successors {
			inDeg[s]--
			if inDeg[s] == 0 {
				queue = append(queue, s)
			}
		}
	}
	if len(result) != len(g.tasks) {
		return nil, fmt.Errorf("cycle detected: only %d of %d tasks orderable", len(result), len(g.tasks))
	}
	return result, nil
}

func (g *Graph) Roots() []string {
	var roots []string
	for _, id := range g.order {
		if len(g.tasks[id].Predecessors) == 0 {
			roots = append(roots, id)
		}
	}
	return roots
}

func (g *Graph) Leaves() []string {
	var leaves []string
	for _, id := range g.order {
		if len(g.tasks[id].Successors) == 0 {
			leaves = append(leaves, id)
		}
	}
	return leaves
}
