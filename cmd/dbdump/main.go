package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/options"
)

func main() {
	opts := badger.DefaultOptions("db").
		WithLogger(nil).
		WithTruncate(true).
		WithValueLogLoadingMode(options.FileIO).
		WithReadOnly(true)

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		fmt.Println("=== BadgerDB Contents ===")
		count := 0

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()

			err := item.Value(func(val []byte) error {
				count++
				fmt.Printf("Key: %s\n", string(key))

				// Try to pretty-print JSON
				var jsonData interface{}
				if err := json.Unmarshal(val, &jsonData); err == nil {
					prettyJSON, _ := json.MarshalIndent(jsonData, "", "  ")
					fmt.Printf("Value: %s\n", string(prettyJSON))
				} else {
					fmt.Printf("Value (raw): %s\n", string(val))
				}
				fmt.Println("---")
				return nil
			})

			if err != nil {
				return err
			}
		}

		fmt.Printf("\nTotal entries: %d\n", count)
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}
