package rdb

import "context"

type RDB interface {
	Shutdown(ctx context.Context)
}
