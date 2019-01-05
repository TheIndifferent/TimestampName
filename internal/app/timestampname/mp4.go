// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"encoding/binary"
	"io"
	"time"
)

var (
	quicktimeEpochOffset = uint32(-time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC).Unix())
)

func mp4ExtractMetadataCreationTimestamp(in reader) string {
	moovIn, err := quicktimeSearchBox(in, "moov")
	CatchFile(err, in.Name(), "moov box not found")
	mvhdIn, err := quicktimeSearchBox(moovIn, "mvhd")
	CatchFile(err, in.Name(), "mvhd box not found")
	var versionBytes = make([]byte, 1)
	_, err = io.ReadFull(mvhdIn, versionBytes)
	var version = versionBytes[0]
	if version > 1 {
		Raise(in.Name(), "unsupported mvhd version")
	}
	var flagBytes = make([]byte, 3)
	_, err = io.ReadFull(mvhdIn, flagBytes)
	if version == 1 {
		var creationTime uint64
		var modificationTime uint64
		err = binary.Read(mvhdIn, binary.BigEndian, &creationTime)
		CatchFile(err, in.Name(), "creation time 64")
		err = binary.Read(mvhdIn, binary.BigEndian, &modificationTime)
		CatchFile(err, in.Name(), "modification time 64")
		var t = int64(modificationTime - uint64(quicktimeEpochOffset))
		return time.Unix(t, 0).UTC().Format("20060102-150405")
	} else {
		var creationTime uint32
		var modificationTime uint32
		err = binary.Read(mvhdIn, binary.BigEndian, &creationTime)
		CatchFile(err, in.Name(), "creation time 32")
		err = binary.Read(mvhdIn, binary.BigEndian, &modificationTime)
		CatchFile(err, in.Name(), "modification time 32")
		var t = int64(modificationTime - quicktimeEpochOffset)
		return time.Unix(t, 0).UTC().Format("20060102-150405")
	}
}
