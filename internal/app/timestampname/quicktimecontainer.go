// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
)

// following documents were used to implement this parser:
// http://l.web.umkc.edu/lizhu/teaching/2016sp.video-communication/ref/mp4.pdf
// https://mpeg.chiariglione.org/standards/mpeg-4/iso-base-media-file-format

func _quicktimeSearchBox(in reader, matchName func(string) bool, matchUuid func(string) bool) reader {
	var err error
	var offset int64              // offset in provided reader
	var boxType = make([]byte, 4) // 4 bytes box type
	for {
		var boxBodyLength int64 // length of the box body
		var boxLength uint32
		err = binary.Read(in, binary.BigEndian, &boxLength)
		CatchFile(err, in.Name(), "failed to read box length")
		_, err = io.ReadFull(in, boxType)
		CatchFile(err, in.Name(), "failed to read box type")
		var boxTypeString = string(boxType)
		debug("quicktime encountered box '%s' at offset %d", boxTypeString, offset)
		// checking for large box:
		if boxLength == 1 {
			var boxLargeLength uint64
			err = binary.Read(in, binary.BigEndian, &boxLargeLength)
			CatchFile(err, in.Name(), "failed to read box large length")
			// box lenght includes header, have to make adjustments:
			// 4 bytes for box length
			// 4 bytes for box type
			// 8 bytes for box large length
			boxBodyLength = int64(boxLargeLength - 16)
			offset += 16
		} else {
			// box lenght includes header, have to make adjustments:
			// 4 bytes for box length
			// 4 bytes for box type
			boxBodyLength = int64(boxLength - 8)
			offset += 8
		}
		if matchName(boxTypeString) {
			if matchUuid == nil {
				debug("quicktime box found at offset: %d, with length: %d", offset, boxBodyLength)
				return newReader(in, offset, boxBodyLength)
			}
			var uuid = make([]byte, 16)
			_, err = io.ReadFull(in, uuid)
			CatchFile(err, in.Name(), "failed to read box uuid")
			// another 16 bytes read:
			boxBodyLength -= 16
			offset += 16
			if matchUuid(hex.EncodeToString(uuid)) {
				debug("quicktime box found at offset: %d, with length: %d", offset, boxBodyLength)
				return newReader(in, offset, boxBodyLength)
			}
		}

		offset += boxBodyLength
		if offset >= in.Size() {
			break // reached the file end
		}
		_, err = in.Seek(offset, 0)
		CatchFile(err, in.Name(), "failed to seek till next box")
	}
	return nil
}

func quicktimeSearchUuidBox(in reader, boxUuidNeeded string) (reader, error) {
	debug("quicktime searching for UUID box: %s", boxUuidNeeded)
	var box = _quicktimeSearchBox(
		in,
		func(name string) bool {
			return name == "uuid"
		},
		func(uuid string) bool {
			return uuid == boxUuidNeeded
		})
	if box == nil {
		return nil, errors.New("failed to find a box with uuid '" + boxUuidNeeded + "'")
	}
	return box, nil
}

func quicktimeSearchBox(in reader, boxTypeNeeded string) (reader, error) {
	debug("quicktime searching for box: %s", boxTypeNeeded)
	var box = _quicktimeSearchBox(
		in,
		func(name string) bool {
			return name == boxTypeNeeded
		},
		nil)
	if box == nil {
		return nil, errors.New("failed to find a box with type '" + boxTypeNeeded + "'")
	}
	return box, nil
}
