package ports

import "context"

type DBHealthChecker interface {
	Check(ctx context.Context) error
}
