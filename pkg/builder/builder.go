package builder

import (
	"context"

	"github.com/daleiii/podsync-web/pkg/feed"
	"github.com/pkg/errors"

	"github.com/daleiii/podsync-web/pkg/model"
)

type Builder interface {
	Build(ctx context.Context, cfg *feed.Config) (*model.Feed, error)
}

func New(ctx context.Context, provider model.Provider, key string, downloader Downloader) (Builder, error) {
	switch provider {
	case model.ProviderYoutube:
		return NewYouTubeBuilder(key, downloader)
	case model.ProviderVimeo:
		return NewVimeoBuilder(ctx, key)
	case model.ProviderSoundcloud:
		return NewSoundcloudBuilder()
	case model.ProviderTwitch:
		return NewTwitchBuilder(key)
	default:
		return nil, errors.Errorf("unsupported provider %q", provider)
	}
}
