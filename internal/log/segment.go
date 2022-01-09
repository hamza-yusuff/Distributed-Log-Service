/*
The segment portion (segment.go) wraps the index struct (defined in the index.go file),
and store types to coordinate operations across the store and index files. T
his is because every time a record gets added to the store file, the index file needs to be updated with the offset and position values.
For reads, the segments needs to look for the index from the index file, and search for the record at that index from the store file.
*/

package log

import (
	"fmt"
	"os"
	"path"

	api "github.com/hamza-yusuff/proglog/api/v1"
	"google.golang.org/protobuf/proto"
)

type segment struct {
	store                  *store
	index                  *index
	baseOffset, nextOffset uint64
	config                 Config
}

func newSegment(dir string, baseOffset uint64, c Config) (*segment, error) {
	// create a segement pointer
	seg := &segment{
		baseOffset: baseOffset,
		config:     c,
	}

	var err error
	// create the store file if not present that's why added OS.O_CREATE FLAG

	storeFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".store")),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)

	if err != nil {
		return nil, err
	}

	// assigns the segement store field with the newstore pointer
	if seg.store, err = newStore(storeFile); err != nil {
		return nil, err
	}

	/// creates the index file if not present

	indexFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".index")),
		os.O_RDWR|os.O_CREATE,
		0644,
	)

	if err != nil {
		return nil, err
	}

	// assigns the segment index field with the new index pointer

	if seg.index, err = newIndex(indexFile, c); err != nil {
		return nil, err
	}

	// setting the baseOffset, if the segment is empty then the next off set would be the baseOffset
	// otherwise for the new offset, the next record should take the offset at the end of segment
	// so the nextOffset would be the summation of baseOffset + 1 + offset of the previous index file
	// for example, if the baseoffset of the index file or segment is 12 (baseOffset value), and if the
	// the last record appended in the segment or index file has an offset of 3(off value), as the index or offset
	//  values are set relatively, then the segment's next offset value would be 12+3+1. Again, to get more
	// clarification go index.go and look at write function to understand how the relative indexing is done.
	if off, _, err := seg.index.Read(-1); err != nil {
		seg.nextOffset = baseOffset
	} else {
		seg.nextOffset = baseOffset + uint64(off) + 1
	}
	return seg, nil

}

// writes the record to the log, and returns the off set of the appended record
// function appends the record to the store, and then adds an index entry
func (seg *segment) Append(record *api.Record) (offset uint64, err error) {
	current := seg.nextOffset
	record.Offset = current

	p, err := proto.Marshal(record)
	if err != nil {
		return 0, err
	}

	_, pos, err := seg.store.Append(p)
	if err != nil {
		return 0, err
	}
	if err = seg.index.Write(uint32(seg.nextOffset-seg.baseOffset), pos); err != nil {
		return 0, err
	}
	seg.nextOffset++
	return current, nil

}

// reads the record at a offset in the index file, that is gets the index entry first
// then with the obtained index entry it goes straight to the record's position in the store

func (seg *segment) Read(off uint64) (*api.Record, error) {

	// first translates the absolute index into relative index
	_, pos, err := seg.index.Read(int64(off - seg.baseOffset))
	if err != nil {
		return nil, err
	}
	p, err := seg.store.Read(pos)
	if err != nil {
		return nil, err
	}
	record := &api.Record{}
	err = proto.Unmarshal(p, record)
	return record, err
}

// returns if the segment has reached its max size or not
// the log uses the method to know if it needs to create a new segment
func (seg *segment) IsMaxed() bool {
	return seg.store.size >= seg.config.Segment.MaxStoreBytes || seg.index.size >= seg.config.Segment.MaxIndexBytes
}

// remove closed the segment, anbd removes the index and store files

func (seg *segment) Remove() error {
	if err := seg.Close(); err != nil {
		return err
	}
	if err := os.Remove(seg.index.Name()); err != nil {
		return err
	}
	if err := os.Remove(seg.store.Name()); err != nil {
		return err
	}
	return nil
}

// to close the segement, that is close the store and index files

func (seg *segment) Close() error {
	if err := seg.index.Close(); err != nil {
		return err
	}
	if err := seg.store.Close(); err != nil {
		return err
	}
	return nil
}

// to see if we have gone pass the user's disk capacity
func nearestMultiple(j, k uint64) uint64 {
	if j >= 0 {
		return (j / k) * k
	}
	return ((j - k + 1) / k) * k
}
