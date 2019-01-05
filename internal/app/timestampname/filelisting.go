// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

type inputFile struct {
	name string
	ext  string
}

func createSupportedFilesMap() map[string]bool {
	sf := make(map[string]bool)
	sf[".dng"] = true
	sf[".nef"] = true
	sf[".jpg"] = true
	sf[".jpeg"] = true
	sf[".mp4"] = true
	sf[".cr3"] = true
	return sf
}

var supportedFiles = createSupportedFilesMap()

func listFiles(targetFolder string) []inputFile {
	files, err := ioutil.ReadDir(targetFolder)
	log.fatalityCheck(err, "error reading contents of current folder %s: %v", targetFolder, err)

	var inputFiles []inputFile
	for _, file := range files {
		// skipping dirs:
		if file.IsDir() {
			continue
		}
		var ext = strings.ToLower(filepath.Ext(file.Name()))
		if !supportedFiles[ext] {
			continue
		}

		var input = inputFile{name: file.Name(), ext: ext}
		inputFiles = append(inputFiles, input)
	}
	return inputFiles
}
