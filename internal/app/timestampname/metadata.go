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
	CatchFile(openErr, file.name, "failed to open")
	defer func() {
		closeErr := openFile.Close()
		CatchFile(closeErr, file.name, "failed to close")
	}()

	var in = newFileReader(openFile, file.name)

	switch file.ext {
	case ".mp4":
		return mp4ExtractMetadataCreationTimestamp(in)
	case ".dng":
		return tiffExtractMetadataCreationTimestamp(in)
	case ".nef":
		return tiffExtractMetadataCreationTimestamp(in)
	case ".jpg":
		return jpegExtractMetadataCreationTimestamp(in)
	case ".jpeg":
		return jpegExtractMetadataCreationTimestamp(in)
	case ".cr3":
		return cr3ExtractMetadataCreationTimestamp(in)
	default:
		Raise(file.name, "unsupported file format")
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
