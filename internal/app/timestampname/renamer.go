// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"fmt"
	"sort"
)

type renameOperation struct {
	from string
	to   string
}

func targetFileNameFormat(numberOfFiles int, noPrefix bool) string {
	if noPrefix {
		return "%[2]s%[3]s"
	}
	if numberOfFiles < 10 {
		return "%d-%s%s"
	}
	if numberOfFiles < 100 {
		return "%02d-%s%s"
	}
	if numberOfFiles < 1000 {
		return "%03d-%s%s"
	}
	if numberOfFiles < 10000 {
		return "%04d-%s%s"
	}
	if numberOfFiles < 100000 {
		return "%05d-%s%s"
	}
	RaiseFmt("too many files: %d", numberOfFiles)
	return ""
}

func prepareRenameOperations(files []fileMetadata, noPrefix bool) ([]renameOperation, int) {
	sort.Slice(files, func(i, j int) bool {
		a := files[i]
		b := files[j]
		if a.metadataCreationTimestamp == b.metadataCreationTimestamp {
			if a.name == b.name {
				Raise(a.name, "encountered twice")
			}
			// workaround for Android way of dealing with same-second shots:
			// 20180430_184327.jpg
			// 20180430_184327(0).jpg
			aLen := len(a.name)
			bLen := len(b.name)
			if aLen == bLen {
				return a.name < b.name
			}
			return aLen < bLen
		}
		return a.metadataCreationTimestamp < b.metadataCreationTimestamp
	})

	var operations = make([]renameOperation, len(files))
	var targetFormat = targetFileNameFormat(len(files), noPrefix)
	var longestSourceName int

	for index, md := range files {
		var targetName = fmt.Sprintf(targetFormat, index+1, md.metadataCreationTimestamp, md.ext)
		operations[index] = renameOperation{md.name, targetName}
		// choosing longest source file name for next operation:
		sourceNameLength := len(md.name)
		if sourceNameLength > longestSourceName {
			longestSourceName = sourceNameLength
		}
	}

	return operations, longestSourceName
}
