// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"encoding/binary"
)

// following resources were used to implement this parser:
// https://www.media.mit.edu/pia/Research/deepview/exif.html
// https://www.fileformat.info/format/jpeg/egff.htm
// http://vip.sugovica.hu/Sardi/kepnezo/JPEG%20File%20Layout%20and%20Format.htm

const (
	jpegSoiExpected          uint16 = 0xFFD8
	jpegApp1                 uint16 = 0xFFE1
	exifHeaderSuffixExpected uint16 = 0x0000
)

var (
	exifHeaderExpected = binary.BigEndian.Uint32([]byte("Exif"))
)

func jpegExtractMetadataCreationTimestamp(in reader) string {
	// checking JPEG SOI:
	var jpegSoi uint16
	binary.Read(in, binary.BigEndian, &jpegSoi)
	if jpegSoi != jpegSoiExpected {
		Raise(in.Name(), "unexpected header")
	}
	// scrolling through fields until we find APP1:
	var offset int64 = 2 // 2 bytes SOI
	for {
		var fieldMarker uint16
		binary.Read(in, binary.BigEndian, &fieldMarker)
		var fieldLength uint16
		binary.Read(in, binary.BigEndian, &fieldLength)
		if fieldMarker == jpegApp1 {
			// APP1 marker found, checking Exif header:
			var exifHeader uint32
			var exifHeaderSuffix uint16
			binary.Read(in, binary.BigEndian, &exifHeader)
			binary.Read(in, binary.BigEndian, &exifHeaderSuffix)
			if exifHeader != exifHeaderExpected || exifHeaderSuffix != exifHeaderSuffixExpected {
				Raise(in.Name(), "JPEG APP1 field does not have valid Exif header")
			}
			// body is a valid TIFF,
			// offset increments:
			//   +2 field marker
			//   +2 field length
			//   +4 exif header
			//   +2 exif header suffix
			// size decrements:
			//   -2 field length
			//   -4 exif header
			//   -2 exif header suffix
			return tiffExtractMetadataCreationTimestamp(newReader(in, offset+10, int64(fieldLength)-8), 0)
		} else {
			// length includes the length itself:
			var scrollDistance = fieldLength - 2
			in.Seek(int64(scrollDistance), 1)
		}
		offset += 2                  // field marker
		offset += int64(fieldLength) // field lenght includes itself
	}
}
