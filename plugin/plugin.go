package plugin

import "context"

// Message is one user message flowing through the agent, tagged with the
// user ID the sending plugin resolved it to (e.g. Telegram chat ID).
type Message struct {
	User string
	Text string
}

type Plugin interface {
	Name() string
	Start(ctx context.Context, in chan<- Message, out <-chan Message) error
}
