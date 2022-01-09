package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	api "github.com/hamza-yusuff/proglog/api/v1"
)

// the log consists of a list of segments and a pointer to the active segment where data is written to
// dir refers to the directory where the segment ( files index and store ) is stored

type Log struct {
	mu            sync.RWMutex
	Config        Config
	Dir           string
	activeSegment *segment
	segments      []*segment
}

// creatng and setting up the log instance

func NewLog(dir string, c Config) (*Log, error) {

	if c.Segment.MaxStoreBytes == 0 {
		c.Segment.MaxStoreBytes = 1024
	}
	if c.Segment.MaxIndexBytes == 0 {
		c.Segment.MaxIndexBytes = 1024
	}
	log := &Log{
		Dir:    dir,
		Config: c,
	}
	return log, log.setup()
}

// setting up the log instance
// when the log starts it sets itselfup for the segemnts on the disk that already exist,
// if there is none, it configures a segment straightup. If present already, the baseoffset
// numbers are parsed from the existing files of segments, then sorted and then new segments
// are created from the baseOffset numbers obatined and sorted. Each offset number correspond
// to a segment on the disk
func (l *Log) setup() error {

	files, err := ioutil.ReadDir(l.Dir)
	if err != nil {
		return err
	}

	var baseOffsets []uint64
	for _, file := range files {
		// removes the extension
		// as the index and store files are stored as baseOffset.store or baseOffset.index,
		offStr := strings.TrimSuffix(
			file.Name(),
			path.Ext(file.Name()),
		)

		off, _ := strconv.ParseUint(offStr, 10, 0)
		baseOffsets = append(baseOffsets, off)
	}
	// sort the baseOffset numbers
	sort.Slice(baseOffsets, func(i, j int) bool {
		return baseOffsets[i] < baseOffsets[j]
	})

	for i := 0; i < len(baseOffsets); i++ {
		// inheritance through embedding
		if err = l.newSegment(baseOffsets[i]); err != nil {
			return err
		}
		// baseOffsets contain duplicates which we safely ignore, this is we parse both index and store files above in the first
		// loop
		i++
	}

	if l.segments == nil {
		if err = l.newSegment(
			l.Config.Segment.InitialOffset,
		); err != nil {
			return err
		}
	}
	return nil
}

// append a log to the active segment, if the segment is maxed out another segement is created
// RWMutex is chosen to grant access to reads when there is not a write holding the lock
func (l *Log) Append(record *api.Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	off, err := l.activeSegment.Append(record)
	if err != nil {
		return 0, err
	}

	// check if the segment is maxed out
	if l.activeSegment.IsMaxed() {
		err = l.newSegment(off + 1)
	}
	return off, err
}

func (l *Log) Read(off uint64) (*api.Record, error) {
	// read locks
	l.mu.RLock()
	defer l.mu.RUnlock()

	var seg *segment

	// look for the segment where the record can be found
	// looked with the off set value
	// iterated until to find the first segment whose base offset is less than or equal to
	// the offset of the record being sought after

	for _, segment := range l.segments {
		if segment.baseOffset <= off && off < segment.nextOffset {
			seg = segment
			break
		}
	}

	if seg == nil || seg.nextOffset <= off {
		return nil, fmt.Errorf("offset out of reange: %d", off)
	}
	return seg.Read(off)

}

// close method iterates over the segmetn and closes them, which in turn closes the index ans store files

func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, segment := range l.segments {
		if err := segment.Close(); err != nil {
			return err
		}
	}
	return nil
}

// remove closes the log, and removes the data

func (l *Log) Remove() error {
	if err := l.Close(); err != nil {
		return err
	}
	return os.RemoveAll(l.Dir)
}

// removes the log, and creates a new log

func (l *Log) Reset() error {
	if err := l.Remove(); err != nil {
		return err
	}

	return l.setup()
}

// Added to support replicated, coordinated cluster

// returns the lowestOffset of the segment
func (l *Log) LowestOffset() (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.segments[0].baseOffset, nil
}

// returns the highest Offset of the segment
func (l *Log) HighestOffset() (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	off := l.segments[len(l.segments)-1].nextOffset
	if off == 0 {
		return 0, nil
	}
	return off - 1, nil
}

// removes all segments whose highest offset is higher than the lowest offset
// this will be called to remove old segments whose does have been processed

func (l *Log) Truncate(lowest uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var segments []*segment

	for _, s := range l.segments {
		if s.nextOffset <= lowest+1 {
			if err := s.Remove(); err != nil {
				return err
			}
			continue
		}
		segments = append(segments, s)
	}

	l.segments = segments
	return nil
}

type originReader struct {
	*store
	off int64
}

func (o *originReader) Read(p []byte) (int, error) {
	n, err := o.ReadAt(p, o.off)
	o.off += int64(n)
	return n, err
}

// Function returns an io.Reader to read the whole logs
// Will allow coordinae consensus, and support for snapshots and restoring of logs
// MultiReader call concatenates the segment stores
func (l *Log) Reader() io.Reader {

	l.mu.RLock()
	defer l.mu.RUnlock()

	readers := make([]io.Reader, len(l.segments))
	for i, segment := range l.segments {
		readers[i] = &originReader{segment.store, 0}
	}
	return io.MultiReader(readers...)
}

// makes a new segments, and assigns that as the active segment for the log
func (l *Log) newSegment(off uint64) error {
	seg, err := newSegment(l.Dir, off, l.Config)

	if err != nil {
		return err
	}

	l.segments = append(l.segments, seg)
	l.activeSegment = seg
	return nil
}
