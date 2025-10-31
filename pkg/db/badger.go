package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/options"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/daleiii/podsync-web/pkg/model"
)

const (
	versionPath   = "podsync/version"
	feedPrefix    = "feed/"
	feedPath      = "feed/%s"
	episodePrefix = "episode/%s/"
	episodePath   = "episode/%s/%s" // FeedID + EpisodeID
	historyPrefix = "history/"
	historyPath   = "history/%s"         // HistoryID (timestamp-uuid)
	historyByFeed = "history_feed/%s/%s" // FeedID + HistoryID
)

// BadgerConfig represents BadgerDB configuration parameters
type BadgerConfig struct {
	Truncate bool `toml:"truncate"`
	FileIO   bool `toml:"file_io"`
}

type Badger struct {
	db *badger.DB
}

var _ Storage = (*Badger)(nil)

func NewBadger(config *Config) (*Badger, error) {
	var (
		dir = config.Dir
	)

	log.Infof("opening database %q", dir)

	// Make sure database directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, errors.Wrap(err, "could not mkdir database dir")
	}

	opts := badger.DefaultOptions(dir).
		WithLogger(log.StandardLogger()).
		WithTruncate(true)

	if config.Badger != nil {
		opts.Truncate = config.Badger.Truncate
		if config.Badger.FileIO {
			opts.ValueLogLoadingMode = options.FileIO
		}
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	storage := &Badger{db: db}

	if err := db.Update(func(txn *badger.Txn) error {
		if err := storage.setObj(txn, []byte(versionPath), CurrentVersion, false); err != nil && err != model.ErrAlreadyExists {
			return err
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to read database version")
	}

	return &Badger{db: db}, nil
}

func (b *Badger) Close() error {
	log.Debug("closing database")
	return b.db.Close()
}

func (b *Badger) Version() (int, error) {
	var (
		version = -1
	)

	err := b.db.View(func(txn *badger.Txn) error {
		return b.getObj(txn, []byte(versionPath), &version)
	})

	return version, err
}

func (b *Badger) AddFeed(_ context.Context, feedID string, feed *model.Feed) error {
	return b.db.Update(func(txn *badger.Txn) error {
		// Insert or update feed info
		feedKey := b.getKey(feedPath, feedID)
		if err := b.setObj(txn, feedKey, feed, true); err != nil {
			return err
		}

		// Append new episodes
		for _, episode := range feed.Episodes {
			episodeKey := b.getKey(episodePath, feedID, episode.ID)
			err := b.setObj(txn, episodeKey, episode, false)
			if err == nil || err == model.ErrAlreadyExists {
				// Do nothing
			} else {
				return errors.Wrapf(err, "failed to save episode %q", feedID)
			}
		}

		return nil
	})
}

func (b *Badger) GetFeed(_ context.Context, feedID string) (*model.Feed, error) {
	var (
		feed    = model.Feed{}
		feedKey = b.getKey(feedPath, feedID)
	)

	if err := b.db.View(func(txn *badger.Txn) error {
		// Query feed
		if err := b.getObj(txn, feedKey, &feed); err != nil {
			return err
		}

		// Set the feed ID from the key parameter
		feed.ID = feedID

		// Query episodes
		if err := b.walkEpisodes(txn, feedID, func(episode *model.Episode) error {
			feed.Episodes = append(feed.Episodes, episode)
			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &feed, nil
}

func (b *Badger) WalkFeeds(_ context.Context, cb func(feed *model.Feed) error) error {
	return b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		prefix := b.getKey(feedPrefix)
		opts.Prefix = prefix
		opts.PrefetchValues = true
		return b.iterator(txn, opts, func(item *badger.Item) error {
			feed := &model.Feed{}
			if err := b.unmarshalObj(item, feed); err != nil {
				return err
			}

			// Extract feed ID from key: podsync/v1/feed/{feedID}
			key := item.Key()
			if len(key) > len(prefix) {
				feed.ID = string(key[len(prefix):])
			}

			return cb(feed)
		})
	})
}

func (b *Badger) DeleteFeed(_ context.Context, feedID string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		// Feed
		feedKey := b.getKey(feedPath, feedID)
		if err := txn.Delete(feedKey); err != nil {
			return errors.Wrapf(err, "failed to delete feed %q", feedID)
		}

		// Episodes
		opts := badger.DefaultIteratorOptions
		opts.Prefix = b.getKey(episodePrefix, feedID)
		opts.PrefetchValues = false
		if err := b.iterator(txn, opts, func(item *badger.Item) error {
			return txn.Delete(item.KeyCopy(nil))
		}); err != nil {
			return errors.Wrapf(err, "failed to iterate episodes for feed %q", feedID)
		}

		return nil
	})
}

func (b *Badger) GetEpisode(_ context.Context, feedID string, episodeID string) (*model.Episode, error) {
	var (
		episode model.Episode
		err     error
		key     = b.getKey(episodePath, feedID, episodeID)
	)

	err = b.db.View(func(txn *badger.Txn) error {
		return b.getObj(txn, key, &episode)
	})

	return &episode, err
}

func (b *Badger) UpdateEpisode(feedID string, episodeID string, cb func(episode *model.Episode) error) error {
	var (
		key     = b.getKey(episodePath, feedID, episodeID)
		episode model.Episode
	)

	return b.db.Update(func(txn *badger.Txn) error {
		if err := b.getObj(txn, key, &episode); err != nil {
			return err
		}

		if err := cb(&episode); err != nil {
			return err
		}

		if episode.ID != episodeID {
			return errors.New("can't change episode ID")
		}

		return b.setObj(txn, key, &episode, true)
	})
}

func (b *Badger) DeleteEpisode(feedID, episodeID string) error {
	key := b.getKey(episodePath, feedID, episodeID)
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (b *Badger) WalkEpisodes(ctx context.Context, feedID string, cb func(episode *model.Episode) error) error {
	return b.db.View(func(txn *badger.Txn) error {
		return b.walkEpisodes(txn, feedID, cb)
	})
}

func (b *Badger) walkEpisodes(txn *badger.Txn, feedID string, cb func(episode *model.Episode) error) error {
	opts := badger.DefaultIteratorOptions
	opts.Prefix = b.getKey(episodePrefix, feedID)
	opts.PrefetchValues = true
	return b.iterator(txn, opts, func(item *badger.Item) error {
		feed := &model.Episode{}
		if err := b.unmarshalObj(item, feed); err != nil {
			return err
		}

		return cb(feed)
	})
}

func (b *Badger) iterator(txn *badger.Txn, opts badger.IteratorOptions, callback func(item *badger.Item) error) error {
	iter := txn.NewIterator(opts)
	defer iter.Close()

	// For reverse iteration with a prefix, we need to seek to the end of the prefix range
	// BadgerDB's Seek positions at the first key >= the seek key
	// For reverse iteration, we want to start at the last key with the prefix
	if opts.Reverse && len(opts.Prefix) > 0 {
		// Create a seek key that's just past the end of the prefix range
		// by appending 0xFF bytes
		seekKey := make([]byte, len(opts.Prefix)+1)
		copy(seekKey, opts.Prefix)
		seekKey[len(opts.Prefix)] = 0xFF
		iter.Seek(seekKey)

		// If we're past the end, rewind to the last item in the prefix
		if !iter.Valid() || !bytes.HasPrefix(iter.Item().Key(), opts.Prefix) {
			iter.Rewind()
		}
	} else {
		iter.Rewind()
	}

	for ; iter.Valid(); iter.Next() {
		item := iter.Item()

		if err := callback(item); err != nil {
			return err
		}
	}

	return nil
}

func (b *Badger) getKey(format string, a ...interface{}) []byte {
	resourcePath := fmt.Sprintf(format, a...)
	fullPath := fmt.Sprintf("podsync/v%d/%s", CurrentVersion, resourcePath)

	return []byte(fullPath)
}

func (b *Badger) setObj(txn *badger.Txn, key []byte, obj interface{}, overwrite bool) error {
	if !overwrite {
		// Overwrites are not allowed, make sure there is no object with the given key
		_, err := txn.Get(key)
		switch err {
		case nil:
			return model.ErrAlreadyExists
		case badger.ErrKeyNotFound:
			// Key not found, do nothing
		default:
			return errors.Wrap(err, "failed to check whether key exists")
		}
	}

	data, err := b.marshalObj(obj)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize object for key %q", key)
	}

	return txn.Set(key, data)
}

func (b *Badger) getObj(txn *badger.Txn, key []byte, out interface{}) error {
	item, err := txn.Get(key)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return model.ErrNotFound
		}

		return err
	}

	return b.unmarshalObj(item, out)
}

func (b *Badger) marshalObj(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

func (b *Badger) unmarshalObj(item *badger.Item, out interface{}) error {
	return item.Value(func(val []byte) error {
		return json.Unmarshal(val, out)
	})
}

// History methods

func (b *Badger) AddHistory(_ context.Context, entry *model.HistoryEntry) error {
	return b.db.Update(func(txn *badger.Txn) error {
		// Store in main history table
		historyKey := b.getKey(historyPath, entry.ID)
		if err := b.setObj(txn, historyKey, entry, true); err != nil {
			return errors.Wrap(err, "failed to save history entry")
		}

		// Store index by feed ID for efficient feed-specific queries
		if entry.FeedID != "" {
			feedIndexKey := b.getKey(historyByFeed, entry.FeedID, entry.ID)
			// Store just the ID as a reference
			if err := txn.Set(feedIndexKey, []byte(entry.ID)); err != nil {
				return errors.Wrap(err, "failed to save feed index")
			}
		}

		return nil
	})
}

func (b *Badger) GetHistory(_ context.Context, id string) (*model.HistoryEntry, error) {
	var (
		entry model.HistoryEntry
		key   = b.getKey(historyPath, id)
	)

	err := b.db.View(func(txn *badger.Txn) error {
		return b.getObj(txn, key, &entry)
	})

	return &entry, err
}

func (b *Badger) ListHistory(_ context.Context, filters model.HistoryFilters, page, pageSize int) ([]*model.HistoryEntry, int, error) {
	var (
		entries []*model.HistoryEntry
		total   int
		skip    = (page - 1) * pageSize
	)

	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true

		// If filtering by feed ID, use the feed index
		if filters.FeedID != "" {
			opts.Prefix = b.getKey(historyByFeed, filters.FeedID, "")
		} else {
			opts.Prefix = b.getKey(historyPrefix)
		}

		// Iterate in reverse order (newest first)
		opts.Reverse = true
		opts.PrefetchValues = true

		return b.iterator(txn, opts, func(item *badger.Item) error {
			entry := &model.HistoryEntry{}

			// If using feed index, we need to lookup the actual entry
			if filters.FeedID != "" {
				// Get the history ID from the index
				var historyID string
				if err := item.Value(func(val []byte) error {
					historyID = string(val)
					return nil
				}); err != nil {
					return err
				}

				// Fetch the actual entry
				historyKey := b.getKey(historyPath, historyID)
				if err := b.getObj(txn, historyKey, entry); err != nil {
					return err
				}
			} else {
				// Direct access from main history table
				if err := b.unmarshalObj(item, entry); err != nil {
					return err
				}
			}

			// Apply filters
			if filters.JobType != "" && entry.JobType != filters.JobType {
				return nil
			}
			if filters.Status != "" && entry.Status != filters.Status {
				return nil
			}
			if !filters.StartDate.IsZero() && entry.StartTime.Before(filters.StartDate) {
				return nil
			}
			if !filters.EndDate.IsZero() && entry.StartTime.After(filters.EndDate) {
				return nil
			}
			if filters.Search != "" && entry.EpisodeTitle != "" {
				// Simple substring search in episode title
				if !contains(entry.EpisodeTitle, filters.Search) {
					return nil
				}
			}

			// Count total matching entries
			total++

			// Skip entries before the current page
			if total <= skip {
				return nil
			}

			// Add to entries
			entries = append(entries, entry)

			// Stop if we've collected enough for this page
			if len(entries) >= pageSize {
				return nil
			}

			return nil
		})
	})

	return entries, total, err
}

func (b *Badger) UpdateHistory(_ context.Context, id string, cb func(entry *model.HistoryEntry) error) error {
	var (
		key   = b.getKey(historyPath, id)
		entry model.HistoryEntry
	)

	return b.db.Update(func(txn *badger.Txn) error {
		if err := b.getObj(txn, key, &entry); err != nil {
			return err
		}

		if err := cb(&entry); err != nil {
			return err
		}

		if entry.ID != id {
			return errors.New("can't change history entry ID")
		}

		return b.setObj(txn, key, &entry, true)
	})
}

func (b *Badger) DeleteHistory(_ context.Context, id string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		// First, get the entry to find the feed ID
		var entry model.HistoryEntry
		key := b.getKey(historyPath, id)
		if err := b.getObj(txn, key, &entry); err != nil {
			if err == model.ErrNotFound {
				return nil // Already deleted
			}
			return err
		}

		// Delete from main history table
		if err := txn.Delete(key); err != nil {
			return errors.Wrap(err, "failed to delete history entry")
		}

		// Delete from feed index if it exists
		if entry.FeedID != "" {
			feedIndexKey := b.getKey(historyByFeed, entry.FeedID, id)
			if err := txn.Delete(feedIndexKey); err != nil {
				// Don't fail if the index doesn't exist
				if err != badger.ErrKeyNotFound {
					return errors.Wrap(err, "failed to delete feed index")
				}
			}
		}

		return nil
	})
}

func (b *Badger) CleanupHistory(_ context.Context, retentionDays int, maxEntries int) error {
	var entriesToDelete []string

	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = b.getKey(historyPrefix)
		opts.PrefetchValues = true
		opts.Reverse = true // Newest first

		var entries []*model.HistoryEntry
		cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

		log.Debugf("CleanupHistory called with retentionDays=%d, maxEntries=%d, prefix=%s", retentionDays, maxEntries, string(opts.Prefix))

		return b.iterator(txn, opts, func(item *badger.Item) error {
			entry := &model.HistoryEntry{}
			if err := b.unmarshalObj(item, entry); err != nil {
				log.WithError(err).Error("failed to unmarshal history entry during cleanup")
				return err
			}

			entries = append(entries, entry)
			log.Debugf("Found history entry: id=%s, feed=%s, time=%s", entry.ID, entry.FeedID, entry.StartTime)

			// Special case: delete all if both retentionDays and maxEntries are 0
			if retentionDays == 0 && maxEntries == 0 {
				log.Debugf("Marking entry %s for deletion (delete all mode)", entry.ID)
				entriesToDelete = append(entriesToDelete, entry.ID)
				return nil
			}

			// Mark for deletion if older than retention period
			if retentionDays > 0 && entry.StartTime.Before(cutoffTime) {
				log.Debugf("Marking entry %s for deletion (older than %d days)", entry.ID, retentionDays)
				entriesToDelete = append(entriesToDelete, entry.ID)
				return nil
			}

			// Mark for deletion if we exceed max entries (keeping newest)
			if maxEntries > 0 && len(entries) > maxEntries {
				log.Debugf("Marking entry %s for deletion (exceeds max entries %d)", entry.ID, maxEntries)
				entriesToDelete = append(entriesToDelete, entry.ID)
			}

			return nil
		})
	})

	if err != nil {
		log.WithError(err).Error("iterator error during cleanup")
		return err
	}

	log.Debugf("CleanupHistory: found %d entries to delete", len(entriesToDelete))

	// Delete marked entries
	for _, id := range entriesToDelete {
		log.Debugf("Deleting history entry: %s", id)
		if err := b.DeleteHistory(context.Background(), id); err != nil {
			log.WithError(err).Errorf("failed to delete history entry %s", id)
			return err
		}
	}

	return nil
}

func (b *Badger) GetHistoryStats(_ context.Context) (count int, oldestEntry *model.HistoryEntry, err error) {
	err = b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = b.getKey(historyPrefix)
		opts.PrefetchValues = true

		var entries []*model.HistoryEntry

		iterErr := b.iterator(txn, opts, func(item *badger.Item) error {
			entry := &model.HistoryEntry{}
			if unmarshalErr := b.unmarshalObj(item, entry); unmarshalErr != nil {
				return unmarshalErr
			}

			entries = append(entries, entry)
			count++

			return nil
		})

		if iterErr != nil {
			return iterErr
		}

		// Find oldest entry
		if len(entries) > 0 {
			oldestEntry = entries[0]
			for _, entry := range entries {
				if entry.StartTime.Before(oldestEntry.StartTime) {
					oldestEntry = entry
				}
			}
		}

		return nil
	})

	return count, oldestEntry, err
}

// Helper function for case-insensitive substring search
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[0] == substr[0] || toLower(s[0]) == toLower(substr[0]))) &&
			(len(s) == 0 || len(substr) == 0 || simpleContains(s, substr)))
}

func simpleContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}
