package plugin

import "context"

type Plugin interface {
	Name() string
	Start(ctx context.Context, in chan<- string, out <-chan string) error
}
