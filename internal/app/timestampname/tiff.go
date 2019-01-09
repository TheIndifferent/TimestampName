// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"encoding/binary"
	"io"
	"sort"
	"time"
)

var (
	tiffEndianessLittle = binary.BigEndian.Uint16([]byte("II"))
	tiffEndianessBig    = binary.BigEndian.Uint16([]byte("MM"))
)

func _removeHead(offsets *[]uint32) {
	var data = *offsets
	if len(data) == 1 {
		*offsets = []uint32{}
	} else {
		data[0] = data[len(data)-1]
		data[len(data)-1] = 0
		*offsets = data[:len(data)-1]
	}
}

// https://www.adobe.io/content/dam/udp/en/open/standards/tiff/TIFF6.pdf
func tiffExtractMetadataCreationTimestamp(in reader) string {
	debug("TIFF processing file: %s", in.Name())
	// Bytes 0-1: The byte order used within the file. Legal values are:
	// “II” (4949.H)
	// “MM” (4D4D.H)
	var tiffEndianess uint16
	// smart thing about specification, we can supplly any endianess:
	err := binary.Read(in, binary.LittleEndian, &tiffEndianess)
	CatchFile(err, in.Name(), "failed to read file header")

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
		RaiseFmtFile(in.Name(), "invalid TIFF file header: %d", tiffEndianess)
	}

	// Bytes 2-3 An arbitrary but carefully chosen number (42)
	// that further identifies the file as a TIFF file.
	var tiffMagic uint16
	err = binary.Read(in, bo, &tiffMagic)
	CatchFile(err, in.Name(), "failed to read TIFF magic number")
	if tiffMagic != 42 {
		RaiseFmtFile(in.Name(), "invalid TIFF magic number: %d", tiffMagic)
	}

	var ifdOffesets = []uint32{0}
	var dateTagOffsets []uint32
	var earliestDate string

	// Bytes 4-7 The offset (in bytes) of the first IFD.
	err = binary.Read(in, bo, &ifdOffesets[0])
	CatchFile(err, in.Name(), "failed to read IFD offset")

	var dateValueBuffer = make([]byte, 19)

	for {
		if len(ifdOffesets) == 0 && len(dateTagOffsets) == 0 {
			debug("TIFF no more offsets to scavenge")
			break // nothing more to collect
		}

		// TODO should sorting happen here?
		// sorting to traverse file forward-only:
		sort.Slice(ifdOffesets, func(i, j int) bool { return ifdOffesets[i] < ifdOffesets[j] })
		sort.Slice(dateTagOffsets, func(i, j int) bool { return dateTagOffsets[i] < dateTagOffsets[j] })

		if len(dateTagOffsets) > 0 || len(ifdOffesets) > 0 {
			var nextDateOffset int64
			var nextIfdOffset int64
			// TODO remove this ugly hack, maybe split big method into submethods:
			if len(dateTagOffsets) > 0 {
				nextDateOffset = int64(dateTagOffsets[0])
			} else {
				var i uint32 = 0
				i--
				nextDateOffset = int64(i)
			}
			if len(ifdOffesets) > 0 {
				nextIfdOffset = int64(ifdOffesets[0])
			} else {
				var i uint32 = 0
				i--
				nextIfdOffset = int64(i)
			}

			if nextDateOffset < nextIfdOffset {
				debug("TIFF collecting date at offset: %d", nextDateOffset)
				_removeHead(&dateTagOffsets)
				// check for overflow, seek position +20 bytes expected field length:
				if nextDateOffset+20 >= in.Size() {
					Raise(in.Name(), "date value offset beyond file length")
				}
				_, err = in.Seek(nextDateOffset, 0)
				CatchFile(err, in.Name(), "failed seeking date tag value")
				_, err = io.ReadFull(in, dateValueBuffer)
				var dateValue = string(dateValueBuffer)
				debug("TIFF date value read: %s", dateValue)
				if len(earliestDate) == 0 {
					earliestDate = dateValue
				} else {
					if dateValue < earliestDate {
						debug("TIFF replacing old value with new value: %s => %s", earliestDate, dateValue)
						earliestDate = dateValue
					}
				}
			} else {
				debug("TIFF scavenging IFD at offset: %d, all offsets: %v", nextIfdOffset, ifdOffesets)
				_removeHead(&ifdOffesets)
				// check for overflow, seek position +2 bytes IFD field count +4 bytes next IFD offset:
				if nextIfdOffset+6 >= in.Size() {
					Raise(in.Name(), "IFD offset goes over file length")
				}
				_, err = in.Seek(nextIfdOffset, 0)
				CatchFile(err, in.Name(), "failed seeking IFD")

				// 2-byte count of the number of directory entries (i.e., the number of fields)
				var fields uint16
				err := binary.Read(in, bo, &fields)
				CatchFile(err, in.Name(), "failed to read number of IFD entries")

				for t := 0; t < int(fields); t++ {
					// Bytes 0-1 The Tag that identifies the field
					var fieldTag uint16
					err := binary.Read(in, bo, &fieldTag)
					CatchFile(err, in.Name(), "failed to read IFD tag")

					// Bytes 2-3 The field Type
					var fieldType uint16
					err = binary.Read(in, bo, &fieldType)
					CatchFile(err, in.Name(), "failed to read IFD type")

					// Bytes 4-7 The number of values, Count of the indicated Type
					var fieldCount uint32
					err = binary.Read(in, bo, &fieldCount)
					CatchFile(err, in.Name(), "failed to read IFD count")

					// Bytes 8-11 The Value Offset, the file offset (in bytes) of the Value for the field
					var fieldValueOffset uint32
					err = binary.Read(in, bo, &fieldValueOffset)
					CatchFile(err, in.Name(), "failed to read IFD value offset")

					// 0x0132: DateTime
					// 0x9003: DateTimeOriginal
					// 0x9004: DateTimeDigitized
					if fieldTag == 0x0132 || fieldTag == 0x9003 || fieldTag == 0x9004 {
						if fieldType != 2 {
							RaiseFmtFile(in.Name(), "expected tag has unexpected type: %d == %d", fieldTag, fieldType)
						}
						if fieldCount != 20 {
							RaiseFmtFile(in.Name(), "expected tag has unexpected size: %d == %d", fieldTag, fieldCount)
						}
						debug("TIFF IFD value offset for tag: %d => %d", fieldTag, fieldValueOffset)
						dateTagOffsets = append(dateTagOffsets, fieldValueOffset)
					}
					// 0x8769: ExifIFDPointer
					if fieldTag == 0x8769 {
						if fieldType != 4 {
							RaiseFmtFile(in.Name(), "EXIF pointer tag has unexpected type: %d == %d", fieldTag, fieldType)
						}
						if fieldCount != 1 {
							RaiseFmtFile(in.Name(), "EXIF pointer tag has unexpected size: %d == %d", fieldTag, fieldCount)
						}
						debug("TIFF IFD Exif offset: %d", fieldValueOffset)
						ifdOffesets = append(ifdOffesets, fieldValueOffset)
					}
				}

				// followed by a 4-byte offset of the next IFD (or 0 if none).
				// (Do not forget to write the 4 bytes of 0 after the last IFD.)
				var parsedIfdOffset uint32
				err = binary.Read(in, bo, &parsedIfdOffset)
				CatchFile(err, in.Name(), "failed to read next IFD offeset")
				debug("TIFF IFD found next IFD offset: %d", parsedIfdOffset)
				if parsedIfdOffset != 0 {
					ifdOffesets = append(ifdOffesets, parsedIfdOffset)
				}
			}
		}
	}

	// fast-forward to the end:
	in.Seek(0, 2)

	parsed, parseError := time.Parse("2006:01:02 15:04:05", earliestDate)
	if parseError != nil {
		// bug in Samsung S9 camera, panorama photo has different date format:
		parsed2, parseError2 := time.Parse("2006-01-02 15:04:05", earliestDate)
		if parseError2 != nil {
			RaiseFmtFile(in.Name(), "failed to parse exif date: %s, %v, %v", earliestDate, parseError, parseError2)
		}
		parsed = parsed2
	}
	return parsed.Format("20060102-150405")
}
