package comparer

import "context"

// Comparer is a type capable of comparing two entities.
type Comparer interface {
	Compare(ctx context.Context)
}
