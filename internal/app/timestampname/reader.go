// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"io"
	"os"
	"reflect"
)

type reader interface {
	ReadAt(p []byte, off int64) (n int, err error)
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
	Size() int64
	Name() string
}

type fileSectionReader struct {
	*io.SectionReader
	name string
}

func (in *fileSectionReader) Name() string {
	return in.name
}

func newFileReader(file *os.File, name string) reader {
	stat, err := file.Stat()
	CatchFile(err, name, "failed to stat")
	return &fileSectionReader{io.NewSectionReader(file, 0, stat.Size()), name}
}

func newReader(r reader, off int64, n int64) reader {
	if cmdArgs.debugOutput {
		var reflectedReader = reflect.ValueOf(r).Elem()
		debug("creating reader: from: {base:%d, off:%d, limit:%d}, new offset: %d, new size: %d",
			reflectedReader.FieldByName("base"),
			reflectedReader.FieldByName("off"),
			reflectedReader.FieldByName("limit"),
			off,
			n)
	}
	return &fileSectionReader{io.NewSectionReader(r, off, n), r.Name()}
}
