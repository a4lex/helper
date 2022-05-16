package myconfig

import (
	"bytes"
	"errors"
	"strings"
)

const (
	KeyDelimiter = "."
	// Node Types
	nodeTypeChildren uint = iota
	nodeTypeScalar
	nodeTypeList
)

var ErrKeyNotFound = errors.New("key not found")
var ErrInvalidKey = errors.New("invalid key")

// node is a config node
type Node struct {
	name     string
	values   []string
	ntype    uint
	children []*Node
	parent   *Node
}

func (n *Node) GetNodeValue(key string) (string, error) {

	list, err := n.GetNodeListValues(key)

	if err != nil {
		return "", err
	}

	if len(list) == 0 {
		return "", ErrKeyNotFound
	}

	return list[0], nil
}

func (n *Node) GetNodeListValues(key string) ([]string, error) {
	node, err := n.getNode(key)

	if err != nil {
		return nil, err
	}

	return node.values, nil
}

func (n *Node) String() string {
	var buff bytes.Buffer
	n.print(&buff, "")
	return buff.String()
}

func (n *Node) getNode(key string) (*Node, error) {

	if key == "" {
		return nil, ErrInvalidKey
	}

	parts := strings.SplitN(key, KeyDelimiter, 2)

	for _, child := range n.children {
		if child.name == parts[0] {
			if len(parts) == 2 {
				return child.getNode(parts[1])
			} else {
				return child, nil
			}
		}
	}

	return nil, ErrKeyNotFound
}

func (n *Node) print(buff *bytes.Buffer, prefix string) {
	buff.WriteString("\n")
	buff.WriteString(prefix)
	buff.WriteString("[")
	buff.WriteString(n.name)
	buff.WriteString("]:")

	for _, val := range n.values {
		buff.WriteString(" ")
		buff.WriteString(val)
	}

	for _, child := range n.children {
		child.print(buff, prefix+"  ")
	}
}

// newChildNode returns the node pointer of the new child created.
func newChildNode(n *Node) *Node {
	newNode := new(Node)
	newNode.parent = n

	n.children = append(n.children, newNode)

	return n.children[len(n.children)-1]
}

// createRootNode returns the node pointer of a "virtual" root node.
func createRootNode() Node {
	rootNode := Node{}
	rootNode.name = "root"
	rootNode.ntype = nodeTypeChildren

	return rootNode
}
