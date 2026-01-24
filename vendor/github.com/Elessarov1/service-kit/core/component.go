package core

import "context"

// Component: Start НЕ должен блокировать.
// Done() сообщает о фатальной ошибке рантайма.
type Component interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Done() <-chan error
}
