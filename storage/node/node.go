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

func CreateRegisterNodeMessage(address string) *Message {
	return &Message{NodeAddress: address, Action: Add}
}

func CreateDeregisterNodeMessage(address string) *Message {
	return &Message{NodeAddress: address, Action: Delete}
}
