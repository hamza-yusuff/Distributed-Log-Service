// code for store file of logs

package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

// enc defines the encoding used for persisting record sizes, and index entries
var (
	enc = binary.BigEndian
)

// number of bytes used for storing record's length, bascillay size of a record struct
const (
	lenWidth = 8
)

// struct to have a pointer to a file, bufio writer, and the size of
type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

// newStore returns a pointer to a new store struct for given file
func newStore(f *os.File) (*store, error) {
	// to get the file's current size
	fi, err := os.Stat(f.Name())

	if err != nil {
		return nil, err
	}

	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil

}

// function appends a p bytes into the buffer by using the buf field in the store struct
// it then returns the position of the written byte on the file, which later gets used
// by the segment of log to append entry (index of the record) in the index file
// position of the record is essentially the size of store file just before appending the record

func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pos = s.size
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}

	written, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}
	written += lenWidth
	s.size += uint64(written)
	return uint64(written), pos, nil
}

// function returns the record stored at the given post
// it first flushed the buffer to the dist
// then reades the record from the file onto an initialized slice of bytes of the required length

func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

// FUnction reads len(p) bytes into p beginnng at the off offset in the stores's file
// implements the io.ReaderAt on store type

func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

// Close Method after ReadAt()
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}
