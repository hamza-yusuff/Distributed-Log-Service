package log

// Implements the methdos and data structures required for reading memory mapped files, in this case index file
import (
	"io"
	"os"

	//"github.com/tysontate/gommap" // allows working with memory mapped files
	"github.com/tysonmote/gommap"
)

// With pos, it means the byte position of the record in the store file
// with offset, it means the index position of the record in the store file.For example, 1,2,3 record index positions.

var (
	// offWidth is the number of bytes used for storing offset value
	// posWidth is the number of bytes used for storing pos value
	// offset values are usually 0,1,2 (record numbers)
	// position values are pos values at the store file
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

// structure has fields to
// to refer to a persisted file
// to refer to a memory maped files, and lastly a size field
//which helps to write the next entry to the index file
type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

// FUnction creates a pointer to an index struct
// it assigns the size field with the size of opened file f
// it then truncates the file to maximum index size by using os.Truncate
// then finally memory maps the files using gommap.Map, and assigns the returned gommap.Map instance to idx.mmp field
func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	idx.size = uint64(fi.Size())
	if err = os.Truncate(
		f.Name(), int64(c.Segment.MaxIndexBytes),
	); err != nil {
		return nil, err
	}

	if idx.mmap, err = gommap.Map(
		idx.file.Fd(),
		gommap.PROT_READ|gommap.PROT_WRITE,
		gommap.MAP_SHARED,
	); err != nil {
		return nil, nil
	}
	return idx, nil

}

/*
Service follows graceful shutdown, and returns the service to a stae where it can restart properly and efficiently
That's why the close method has logic to truncate the persisted file first, by removing the empty spaces between
the last record in the index file and the end of file which was there before to compute the maximum possible file size ( Open function did that)
By truncating the persisted file, we remove the empty spaces, and make sure the last entry is the last record appened in the file, and is
at the end of file. With space inside the file between the last record and end of file, the service could not restart properly but truncating
can make the graceful shutdown, and restart possible eventually.

*/

// Function closes the file stored at i.file
// it first makes sure the memory mapped file at idx.mmap is synced to the actual file or not
// it then truncates the persisted file to the amount of data that's actually in it
func (i *index) Close() error {
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}
	if err := i.file.Sync(); err != nil {
		return err
	}
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close()

}

// takes an offset and returns associated record's position in store
// Offset values are usually 0,1. So, position would be computed
// as offset value * entWidth.
func (i *index) Read(in int64) (off uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	if in == -1 {
		off = uint32((i.size / entWidth) - 1)
	} else {
		off = uint32(in)
	}
	pos = uint64(off) * entWidth
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}
	off = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return off, pos, nil
}

// Write function appends the given offset and position to the index. We
// first validate if we have space to write, if we find there is space, we
// encode the offset and position values to ultimately use them to write to the memory mapped files

func (i *index) Write(off uint32, pos uint64) error {
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	i.size += uint64(entWidth)
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}
