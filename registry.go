package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// FileOffset store file offset and last update time.
type FileOffset struct {
	ID     string
	Offset int64
	Time   int64
}

// Registry manage FileOffsets.
type Registry struct {
	path    string
	ttl     int64
	offsets map[string]FileOffset
}

// NewRegistry return a Registry with offset ttl.
func NewRegistry(path string, ttl int64) Registry {
	return Registry{
		path:    path,
		ttl:     ttl,
		offsets: make(map[string]FileOffset),
	}
}

// Get get FileOffset from registry.
func (r Registry) Get(id string) (FileOffset, bool) {
	v, ok := r.offsets[id]
	return v, ok
}

// Remove remove FileOffset from registry.
func (r Registry) Remove(id string) {
	delete(r.offsets, id)
}

// Set add FileOffset to registry.
func (r Registry) Set(v FileOffset) {
	r.offsets[v.ID] = v
}

// Load read registry from file.
func (r Registry) Load() error {
	// file not exists
	if _, err := os.Stat(r.path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	file, err := os.Open(r.path)
	if err != nil {
		return err
	}
	defer file.Close()

	now := time.Now().Unix()
	s := bufio.NewScanner(file)
	for s.Scan() {
		line := strings.Split(s.Text(), ",")
		id := line[0]
		offset, err := strconv.ParseInt(line[1], 10, 64)
		if err != nil {
			return err
		}
		timestamp, err := strconv.ParseInt(line[2], 10, 64)
		if err != nil {
			return err
		}
		if now-timestamp < r.ttl {
			r.offsets[id] = FileOffset{
				ID:     id,
				Offset: offset,
				Time:   timestamp,
			}
		}
	}
	return s.Err()
}

// Dump dump registry to file.
func (r Registry) Dump() error {
	tmp := r.path + ".1"
	file, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	now := time.Now().Unix()
	for _, v := range r.offsets {
		if now-v.Time > r.ttl {
			r.Remove(v.ID)
			continue
		}
		_, err := fmt.Fprintf(file, "%s,%d,%d\n", v.ID, v.Offset, v.Time)
		if err != nil {
			file.Close()
			return err
		}
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, r.path)
}
