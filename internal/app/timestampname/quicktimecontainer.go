// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
)

func _quicktimeSearchBox(in reader, matchName func(string) bool, matchUuid func(string) bool) reader {
	var err error
	var offset int64              // offset in provided reader
	var boxType = make([]byte, 4) // 4 bytes box type
	for {
		var boxBodyLength int64 // length of the box body
		var boxLength uint32
		err = binary.Read(in, binary.BigEndian, &boxLength)
		log.fatalityCheck(err, "failed to read box length: %v", err)
		_, err = io.ReadFull(in, boxType)
		log.fatalityCheck(err, "failed to read box type: %v")
		// checking for large box:
		if boxLength == 1 {
			var boxLargeLength uint64
			err = binary.Read(in, binary.BigEndian, &boxLargeLength)
			log.fatalityCheck(err, "failed to read box large length: %v", err)
			// 4 bytes for box length
			// 4 bytes for box type
			// 8 bytes for box large length
			boxBodyLength = int64(boxLargeLength - 16)
			offset += 16
		} else {
			// 4 bytes for box length
			// 4 bytes for box type
			boxBodyLength = int64(boxLength - 8)
			offset += 8
		}
		if matchName(string(boxType)) {
			if matchUuid == nil {
				return newReader(in, offset, boxBodyLength)
			}
			var uuid = make([]byte, 16)
			_, err = io.ReadFull(in, uuid)
			log.fatalityCheck(err, "failed to read box uuid: %s, %v", in.Name(), err)
			// another 16 bytes read:
			boxBodyLength -= 16
			offset += 16
			if matchUuid(hex.EncodeToString(uuid)) {
				return newReader(in, offset, boxBodyLength)
			}
		}

		offset += boxBodyLength
		if offset >= in.Size() {
			break // reached the file end
		}
		_, err = in.Seek(offset, 0)
		log.fatalityCheck(err, "failed to seek till next box: %s, %v", in.Name(), err)
	}
	return nil
}

func quicktimeSearchUuidBox(in reader, boxUuidNeeded string) (reader, error) {
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
	// var err error
	// var offset int64              // offset in provided reader
	// var boxType = make([]byte, 4) // 4 bytes box type
	// for {
	// 	var boxBodyLength int64 // length of the box body
	// 	var boxLength uint32
	// 	err = binary.Read(in, binary.BigEndian, &boxLength)
	// 	log.fatalityCheck(err, "failed to read box length: %v", err)
	// 	_, err = io.ReadFull(in, boxType)
	// 	log.fatalityCheck(err, "failed to read box type: %v")
	// 	// checking for large box:
	// 	if boxLength == 1 {
	// 		var boxLargeLength uint64
	// 		err = binary.Read(in, binary.BigEndian, &boxLargeLength)
	// 		log.fatalityCheck(err, "failed to read box large length: %v", err)
	// 		// 4 bytes for box length
	// 		// 4 bytes for box type
	// 		// 8 bytes for box large length
	// 		boxBodyLength = int64(boxLargeLength - 16)
	// 		offset += 16
	// 	} else {
	// 		// 4 bytes for box length
	// 		// 4 bytes for box type
	// 		boxBodyLength = int64(boxLength - 8)
	// 		offset += 8
	// 	}
	// 	if string(boxType) == boxTypeNeeded {
	// 		return newReader(in, offset, boxBodyLength), nil
	// 	}

	// 	offset += boxBodyLength
	// 	if offset >= in.Size() {
	// 		break // reached the file end
	// 	}
	// 	_, err = in.Seek(offset, 0)
	// 	log.fatalityCheck(err, "failed to seek till next box: %s, %v", in.Name(), err)
	// }
	// return nil, nil
}
