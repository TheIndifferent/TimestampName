// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"os"
)

type fileMetadata struct {
	inputFile
	metadataCreationTimestamp string
}

func extractMetadataCreationTimestamp(file inputFile) string {

	openFile, openErr := os.Open(file.name)
	log.fatalityCheck(openErr, "failed to open the file: %s, %v", file.name, openErr)
	defer func() {
		closeErr := openFile.Close()
		log.fatalityCheck(closeErr, "failed to close the file: %s, %v", file.name, closeErr)
	}()

	var in = newFileReader(openFile, file.name)
	fileStat, err := openFile.Stat()
	log.fatalityCheck(err, "failed to stat the file: %s, %v", file.name, err)
	var fileSize = uint32(fileStat.Size())

	switch file.ext {
	case ".mp4":
		return mp4ExtractMetadataCreationTimestamp(file, openFile, fileSize)
	case ".dng":
		return tiffExtractMetadataCreationTimestamp(in, 0)
	case ".nef":
		return tiffExtractMetadataCreationTimestamp(in, 0)
	case ".jpg":
		return jpegExtractMetadataCreationTimestamp(in)
	case ".jpeg":
		return jpegExtractMetadataCreationTimestamp(in)
	case ".cr3":
		return cr3ExtractMetadataCreationTimestamp(in)
	default:
		log.fatalityDo("Unsupported file format: %s", file.ext)
		return ""
	}
}

func fileMetadataCreationTimestamp(file inputFile) fileMetadata {
	var metadataCreationTimestamp = extractMetadataCreationTimestamp(file)
	var metadata = fileMetadata{
		inputFile:                 file,
		metadataCreationTimestamp: metadataCreationTimestamp}
	return metadata
}
