package server

// Simple Commit Log service to faciliate a structure called Log where logs are appended
// as record structure (defined in the code) inside a slice of records
// Methods Append and Read faciliate appending data and reading from the defined structures

import (
	"fmt"
	"sync"
)

var ErrOffsetNotFound = fmt.Errorf("offset not found")

type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

type Log struct {
	mu      sync.Mutex
	records []Record
}

// returns a pointer to the Log structure
func NewLog() *Log {
	return &Log{}
}

// Function is an assigned method of the structure Log
// It takes a record struct and appends it to the Log structure's
// records slice. Returns the old length of record slice in log, and error object
func (log *Log) Append(record Record) (uint64, error) {
	log.mu.Lock()
	defer log.mu.Unlock()

	record.Offset = uint64(len(log.records))
	log.records = append(log.records, record)
	return record.Offset, nil
}

// Function uses the offset value as index for the slice of
// of records in Log. Returns the record at the index offset, and an error object
func (log *Log) Read(offset uint64) (Record, error) {
	log.mu.Lock()
	defer log.mu.Unlock()
	if offset >= uint64(len(log.records)) {
		return Record{}, ErrOffsetNotFound
	}
	return log.records[offset], nil
}
