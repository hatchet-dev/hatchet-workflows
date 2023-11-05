package types

type WorkflowTree struct {
	RootWorkflows []*WorkflowNode
}

type WorkflowNode struct {
	Name     string
	Children []*WorkflowNode
}

func NewWorkflowTree() *WorkflowTree {
	return &WorkflowTree{
		RootWorkflows: []*WorkflowNode{},
	}
}

func newNode(name string) *WorkflowNode {
	return &WorkflowNode{
		Name:     name,
		Children: []*WorkflowNode{},
	}
}

func (t *WorkflowTree) AddRootNode(name string) {
	t.RootWorkflows = append(t.RootWorkflows, newNode(name))
}

func (t *WorkflowNode) AddNode(name string) error {
	t.Children = append(t.Children, newNode(name))

	return nil
}

func ParseWorkflowTreeFromFile(file WorkflowFile) (*WorkflowTree, error) {
	tree := NewWorkflowTree()

	// add all jobs as root nodes. when dependencies are supported this will change.
	for jobName := range file.Jobs {
		tree.AddRootNode(jobName)
	}

	return tree, nil
}
