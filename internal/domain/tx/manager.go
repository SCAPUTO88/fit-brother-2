package tx

import "context"

// Manager abstrai transações atômicas mantendo a camada de aplicação independente de SQL.
type Manager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
