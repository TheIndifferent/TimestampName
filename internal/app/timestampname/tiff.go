// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"encoding/binary"
	"sort"
	"time"
)

var (
	tiffEndianessLittle = binary.BigEndian.Uint16([]byte("II"))
	tiffEndianessBig    = binary.BigEndian.Uint16([]byte("MM"))
)

func tiffAppendDateValueOffsetsFromIFD(in reader, bo binary.ByteOrder, dateTagOffsets []uint32) ([]uint32, uint32) {
	// 2-byte count of the number of directory entries (i.e., the number of fields)
	var fields uint16
	err := binary.Read(in, bo, &fields)
	log.fatalityCheck(err, "failed to read number of IFD entries: %s, %v", in.Name(), err)

	// EXIF IFD will be needed after parsing all current IFDs:
	var exifOffset uint32

	for t := 0; t < int(fields); t++ {
		// Bytes 0-1 The Tag that identifies the field
		var fieldTag uint16
		err := binary.Read(in, bo, &fieldTag)
		log.fatalityCheck(err, "failed to read IFD tag: %s, %v", in.Name(), err)

		// Bytes 2-3 The field Type
		var fieldType uint16
		err = binary.Read(in, bo, &fieldType)
		log.fatalityCheck(err, "failed to read IFD type: %s, %v", in.Name(), err)

		// Bytes 4-7 The number of values, Count of the indicated Type
		var fieldCount uint32
		err = binary.Read(in, bo, &fieldCount)
		log.fatalityCheck(err, "failed to read IFD count: %s, %v", in.Name(), err)

		// Bytes 8-11 The Value Offset, the file offset (in bytes) of the Value for the field
		var fieldValueOffset uint32
		err = binary.Read(in, bo, &fieldValueOffset)
		log.fatalityCheck(err, "failed to read IFD value offset: %s, %v", in.Name(), err)

		// 0x0132: DateTime
		// 0x9003: DateTimeOriginal
		// 0x9004: DateTimeDigitized
		if fieldTag == 0x0132 || fieldTag == 0x9003 || fieldTag == 0x9004 {
			if fieldType != 2 {
				log.fatalityDo("expected tag has unexpected type in file %s: %d == %d", in.Name(), fieldTag, fieldType)
			}
			if fieldCount != 20 {
				log.fatalityDo("expected tag has unexpected size in file %s: %d == %d", in.Name(), fieldTag, fieldCount)
			}
			dateTagOffsets = append(dateTagOffsets, fieldValueOffset)
		}

		// 0x8769: ExifIFDPointer
		if fieldTag == 0x8769 {
			if fieldType != 4 {
				log.fatalityDo("EXIF pointer tag has unexpected type in file %s: %d == %d", in.Name(), fieldTag, fieldType)
			}
			if fieldCount != 1 {
				log.fatalityDo("EXIF pointer tag has unexpected size in file %s: %d == %d", in.Name(), fieldTag, fieldCount)
			}
			exifOffset = fieldValueOffset
		}
	}

	return dateTagOffsets, exifOffset
}

// https://www.adobe.io/content/dam/udp/en/open/standards/tiff/TIFF6.pdf
func tiffExtractMetadataCreationTimestamp(in reader, fileStartOffset int64) string {
	// Bytes 0-1: The byte order used within the file. Legal values are:
	// “II” (4949.H)
	// “MM” (4D4D.H)
	var tiffEndianess uint16
	// smart thing about specification, we can supplly any endianess:
	err := binary.Read(in, binary.LittleEndian, &tiffEndianess)
	log.fatalityCheck(err, "failed to read file header: %s, %v", in.Name(), err)

	// In the “II” format, byte order is always from the least significant byte to the most
	// significant byte, for both 16-bit and 32-bit integers.
	// This is called little-endian byte order.
	//  In the “MM” format, byte order is always from most significant to least
	// significant, for both 16-bit and 32-bit integers.
	// This is called big-endian byte order
	var bo binary.ByteOrder
	switch tiffEndianess {
	case tiffEndianessBig:
		bo = binary.BigEndian
	case tiffEndianessLittle:
		bo = binary.LittleEndian
	default:
		log.fatalityDo("invalid TIFF file header for file %s: %v", in.Name(), tiffEndianess)
	}

	// Bytes 2-3 An arbitrary but carefully chosen number (42)
	// that further identifies the file as a TIFF file.
	var tiffMagic uint16
	err = binary.Read(in, bo, &tiffMagic)
	log.fatalityCheck(err, "failed to read TIFF magic number: %s, %v", in.Name(), err)
	if tiffMagic != 42 {
		log.fatalityDo("invalid TIFF magic number %s: %v", in.Name(), tiffMagic)
	}

	// Bytes 4-7 The offset (in bytes) of the first IFD.
	var ifdOffset uint32
	err = binary.Read(in, bo, &ifdOffset)
	log.fatalityCheck(err, "failed to read IFD offset: %s, %v", in.Name(), err)

	// offsets for date tags we are looking for:
	var dateTagOffsets []uint32
	// offset for EXIF IFD:
	var exifOffset uint32

	// saving previous offset to protect against recursive IFD:
	var ifdOffsetPrev = ifdOffset
	for ifdOffset != 0 {
		// check for overflow, seek position +2 bytes IFD field count +4 bytes next IFD offset:
		if int64(ifdOffset)+6 >= in.Size() {
			log.fatalityDo("IFD offset goes over file length: %d, %s", ifdOffset, in.Name())
		}

		// seek the IFD:
		_, err := in.Seek(fileStartOffset+int64(ifdOffset), 0)
		log.fatalityCheck(err, "failed to seek IFD offset: %s, %v", in.Name(), err)

		var eo uint32
		dateTagOffsets, eo = tiffAppendDateValueOffsetsFromIFD(in, bo, dateTagOffsets)
		if eo != 0 {
			if exifOffset != 0 && eo != exifOffset {
				log.fatalityDo("found multiple Exif offsets inside IFD offset %d: %d, %d", ifdOffset, exifOffset, eo)
			}
			exifOffset = eo
		}

		// TODO check what happens without it
		// we are looking for only 3 tags:
		if len(dateTagOffsets) == 3 {
			break
		}

		err = binary.Read(in, bo, &ifdOffset)
		log.fatalityCheck(err, "failed to read next IFD offeset: %s, %v", in.Name(), err)

		if ifdOffset == ifdOffsetPrev {
			log.fatalityDo("recursive IFD is not supported: %s", in.Name())
		}
		// if EXIF offset matches next offset then skip separate EXIF reading:
		if ifdOffset == exifOffset {
			exifOffset = 0
		}
		ifdOffsetPrev = ifdOffset
	}
	// read EXIF IFD:
	for exifOffset != 0 {
		// check for overflow, seek position +2 bytes IFD field count:
		if int64(exifOffset)+2 >= in.Size() {
			log.fatalityDo("Exif offset goes over file length: %d, %s", exifOffset, in.Name())
		}
		_, err = in.Seek(fileStartOffset+int64(exifOffset), 0)
		log.fatalityCheck(err, "failed to seek EXIF offset: %s, %v", in.Name(), err)
		var exifOffsetPrev = exifOffset
		dateTagOffsets, exifOffset = tiffAppendDateValueOffsetsFromIFD(in, bo, dateTagOffsets)
		// protection from recursive offsets:
		if exifOffset == exifOffsetPrev {
			break
		}
	}

	if len(dateTagOffsets) == 0 {
		log.fatalityDo("no date tags found in file: %s", in.Name())
	}
	// sort to read from closest tag:
	sort.Slice(dateTagOffsets, func(i, j int) bool {
		return dateTagOffsets[i] < dateTagOffsets[j]
	})

	// 2 = ASCII 8-bit byte that contains a 7-bit ASCII code; the last byte must be NUL (binary zero).
	// tag count is 20, which means 19 chars and binary NUL,
	// we will read only 19 bytes then:
	var earliestDate string
	var buffer = make([]byte, 19)
	for _, tagOffset := range dateTagOffsets {
		// check for overflow, seek position +20 bytes expected field length:
		if int64(tagOffset)+20 >= in.Size() {
			log.fatalityDo("Date value offset goes over file length: %d, %s", tagOffset, in.Name())
		}
		_, err = in.Seek(fileStartOffset+int64(tagOffset), 0)
		log.fatalityCheck(err, "failed seeking date tag value: %s, %v", in.Name(), err)
		_, err = in.Read(buffer)
		log.fatalityCheck(err, "failed to read date tag value: %s, %v", in.Name(), err)
		if len(earliestDate) == 0 {
			earliestDate = string(buffer)
		} else {
			str := string(buffer)
			if str < earliestDate {
				earliestDate = str
			}
		}
	}

	parsed, parseError := time.Parse("2006:01:02 15:04:05", earliestDate)
	if parseError != nil {
		// bug in Samsung S9 camera, panorama photo has different date format:
		parsed2, parseError2 := time.Parse("2006-01-02 15:04:05", earliestDate)
		if parseError2 != nil {
			log.fatalityDo("failed to parse exif date: %s\n\t%v\n\t%v")
		}
		parsed = parsed2
	}
	return parsed.Format("20060102-150405")
}
