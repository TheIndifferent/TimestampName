// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

func mp4ExtractMetadataCreationTimestamp(fileInfo inputFile, file *os.File, fileSize uint32) string {
	var buffer = make([]byte, 8)
	var processed uint32
	for {
		// next atom:
		_, err := io.ReadFull(file, buffer)
		log.fatalityCheck(err, "failed to read the file: %s, %v", fileInfo.name, err)
		var atomLen = binary.BigEndian.Uint32(buffer[:4])
		var atomTypeStr = string(buffer[4:])
		if atomTypeStr == "moov" {
			var processedMoov uint32 = 8
			for {
				// next sub-atom:
				_, err = io.ReadFull(file, buffer)
				log.fatalityCheck(err, "failed to read the file: %s, %v", fileInfo.name, err)
				var moovSubAtomLen = binary.BigEndian.Uint32(buffer[:4])
				var moovSubAtomTypeStr = string(buffer[4:])
				if moovSubAtomTypeStr == "mvhd" {
					_, err = io.ReadFull(file, buffer)
					log.fatalityCheck(err, "failed to read the file: %s, %v", fileInfo.name, err)
					version := buffer[0]
					var seconds uint64
					_, err = io.ReadFull(file, buffer)
					log.fatalityCheck(err, "failed to read the file: %s, %v", fileInfo.name, err)
					if version == 1 {
						seconds = binary.BigEndian.Uint64(buffer)
					} else {
						seconds = uint64(binary.BigEndian.Uint32(buffer[:4]))
					}
					// mp4 starts from year 1904:
					offset := uint64(time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC).Unix())
					// adding negative offset:
					seconds += offset
					return time.Unix(int64(seconds), 0).UTC().Format("20060102-150405")
				} else {
					processedMoov += moovSubAtomLen
					processed += moovSubAtomLen
				}
				if processedMoov >= atomLen {
					break
				}
				file.Seek(int64(processed), 0)
			}
		} else {
			processed += atomLen
		}
		if processed >= fileSize {
			break
		}
		file.Seek(int64(processed), 0)
	}
	fmt.Fprintf(os.Stderr, "movie header not found in file: %s\n", fileInfo.name)
	os.Exit(1)
	return ""
}
