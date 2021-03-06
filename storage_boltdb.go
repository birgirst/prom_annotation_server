package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

type BoltDBStorage struct {
	seqVal int
	fName  string
	db     *bolt.DB
}

func NewBoltDBStorage(n string) (*BoltDBStorage, error) {
	db, err := bolt.Open(n, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &BoltDBStorage{db: db, fName: n}, err
}

func (s *BoltDBStorage) seq() int {
	s.seqVal++
	return s.seqVal
}

func (s *BoltDBStorage) TagStats() (res TagStats, err error) {
	res = make(map[string]int)
	s.db.View(func(tx *bolt.Tx) error {
		tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			stats := b.Stats()
			res[string(name)] += stats.KeyN
			return nil
		})
		return nil
	})

	return res, nil
}

func (s *BoltDBStorage) AllTags() (res []string) {
	res = []string{}
	s.db.View(func(tx *bolt.Tx) error {
		tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			res = append(res, string(name))
			return nil
		})
		return nil
	})

	return res
}

func (s *BoltDBStorage) Add(a Annotation) error {
	// make a copy of a and skip the tags, we don't need them in the DB
	val, _ := json.Marshal(Annotation{CreatedAt: a.CreatedAt, Message: a.Message})

	err := s.db.Update(func(tx *bolt.Tx) error {
		for _, tag := range a.Tags {
			b, err := tx.CreateBucketIfNotExists([]byte(tag))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
			key := fmt.Sprintf("%s-seq:%d", time.Unix(int64(a.CreatedAt), 0).Format(time.RFC3339), s.seq())
			if err = b.Put([]byte(key), val); err != nil {
				return fmt.Errorf("err adding to bucket: %s", err)
			}
		}
		return nil
	})

	return err
}

func (s *BoltDBStorage) GetCount(tag string) (count int) {
	s.db.View(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket([]byte(tag))
		if b == nil {
			return
		}
		count = b.Stats().KeyN
		return
	})
	return
}

func (s *BoltDBStorage) ListForTag(tag string, r, until int, out *[]Annotation) (err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tag))
		if b == nil {
			return nil
		}

		start := []byte(time.Unix(int64(until-r), 0).Format(time.RFC3339))
		end := []byte(time.Unix(int64(until), 0).Format(time.RFC3339))

		c := b.Cursor()
		for k, v := c.Seek(start); k != nil && bytes.Compare(k[:len(end)], end) <= 0; k, v = c.Next() {
			var a Annotation
			if err := json.Unmarshal(v, &a); err != nil {
				return err
			}
			*out = append(*out, Annotation{CreatedAt: a.CreatedAt * 1000, Message: a.Message, Tags: []string{tag}})
		}
		return nil
	})
	return
}

func (s *BoltDBStorage) Close() {
	s.db.Close()
	log.Printf("Closed BoltDB storage")
}

func (s *BoltDBStorage) Cleanup() {
	s.Close()
	os.Remove(s.fName)
}
