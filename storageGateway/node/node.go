package node

type ActionType uint

const (
	Add    ActionType = iota
	Delete ActionType = iota
)

type Message struct {
	NodeAddress string
	Action      ActionType
}
