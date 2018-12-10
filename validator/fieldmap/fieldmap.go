package fieldmap

import "github.com/vektah/gqlparser/ast"

type Map struct {
	keys   []string
	values [][]*Node
	lookup map[string]int
}

type Node struct {
	Field *ast.Field
	Path  Path
}

func (m *Map) Push(alias string, node *Node) {
	// make a copy of the node path
	pathcopy := make([]interface{}, len(node.Path))
	for i, v := range node.Path {
		pathcopy[i] = v
	}
	node.Path = pathcopy

	if m.lookup == nil {
		m.lookup = map[string]int{}
	}

	if i, ok := m.lookup[alias]; ok {
		m.values[i] = append(m.values[i], node)
		return
	}

	m.lookup[alias] = len(m.keys)
	m.keys = append(m.keys, alias)
	m.values = append(m.values, []*Node{node})
}

func (m *Map) Get(alias string) ([]*Node, bool) {
	if m.lookup == nil {
		return nil, false
	}
	if i, ok := m.lookup[alias]; ok {
		return m.values[i], true
	}
	return nil, false
}

func (m *Map) Len() int {
	return len(m.keys)
}

func (m *Map) At(i int) (key string, value []*Node) {
	return m.keys[i], m.values[i]
}
func (m *Map) Reset() {
	m.keys = m.keys[0:0]
	m.values = m.values[0:0]
	m.lookup = nil
}
