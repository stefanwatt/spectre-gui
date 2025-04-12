package undo

type ReplaceOp struct {
	Path         string
	OriginalText string
	Row          int
}
type ReplaceAction struct {
	Actions []ReplaceOp
}

type Node[T any] struct {
	value    T
	previous *Node[T]
}

type Stack[T any] interface {
	Pop() *Node[T]
	Push(value T)
}

type UndoStack struct {
	head *Node[ReplaceAction]
}

func (s *UndoStack) Pop() ReplaceAction {
	if s.head == nil {
		panic("Cannot pop from an empty stack")
	}
	current_head := s.head
	s.head = current_head.previous
	return current_head.value
}

func (s *UndoStack) Push(value ReplaceAction) {
	new_node := Node[ReplaceAction]{
		value:    value,
		previous: s.head,
	}
	s.head = &new_node
}

func (s *UndoStack) IsEmpty() bool {
	return s.head == nil
}
