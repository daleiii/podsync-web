//go:generate mockgen -source=deps.go -destination=deps_mock_test.go -package=feed

package feed

import (
	"context"

	"github.com/daleiii/podsync-web/pkg/model"
)

type feedProvider interface {
	GetFeed(ctx context.Context, feedID string) (*model.Feed, error)
}
