// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"flag"
	"os"
)

//
// BEFORE INITIALIZATION
//

type commandLineArguments struct {
	dryRun      bool
	noPrefix    bool
	debugOutput bool
}

func parseCommandLineArguments() commandLineArguments {
	var cmdArgs commandLineArguments
	flag.BoolVar(&cmdArgs.dryRun, "dry", false, "dry run")
	flag.BoolVar(&cmdArgs.noPrefix, "noprefix", false, "no counter prefix")
	flag.BoolVar(&cmdArgs.debugOutput, "debug", false, "debug output")
	flag.Parse()
	return cmdArgs
}

//
// END BEFORE INITIALIZATION
//

var log logger
var workDir string

func processFiles(files []inputFile) []fileMetadata {
	var total = len(files)
	var output = make([]fileMetadata, total)
	for index, file := range files {
		log.info("\rProcessing files: %d/%d...", index+1, total)
		output[index] = fileMetadataCreationTimestamp(file)
	}
	log.info(" done.\n")
	return output
}

func verifyOperations(operations []renameOperation, longestSourceName int) {
	duplicatesMap := make(map[string]string)
	for _, operation := range operations {
		log.info("    %[3]*[1]s    =>    %[2]s\n", operation.from, operation.to, longestSourceName)
		// check for target name duplicates:
		if _, existsInMap := duplicatesMap[operation.to]; existsInMap {
			log.fatalityDo("target file name duplicate: %s", operation.to)
		} else {
			duplicatesMap[operation.to] = operation.to
		}
		// check for renaming duplicates:
		if operation.from != operation.to {
			if _, existsInDir := os.Stat(operation.to); existsInDir == nil {
				log.fatalityDo("target file exists on the file system: %s", operation.to)
			}
		}
	}
}

func executeOperations(operations []renameOperation, dryRun bool) {
	for index, operation := range operations {
		log.info("\rRenaming files: %d/%d", index+1, len(operations))
		if !dryRun {
			renameErr := os.Rename(operation.from, operation.to)
			log.fatalityCheck(renameErr, "failed to rename file: %s => %s, %v", operation.from, operation.to, renameErr)
			chmodErr := os.Chmod(operation.to, 0444)
			log.fatalityCheck(chmodErr, "failed to change permissions for file: %s, %v", operation.to, chmodErr)
		}
	}
	log.info(" done.\n")
}

func Exec() {
	var cmdArgs = parseCommandLineArguments()
	log = newLog(cmdArgs.debugOutput)

	log.info("Scanning for files... ")
	var err error
	workDir, err = os.Getwd()
	log.fatalityCheck(err, "failed to get current working directory: %v", err)
	var inputFiles = listFiles(workDir)
	log.info("%d supported files found.\n", len(inputFiles))

	metadatas := processFiles(inputFiles)
	log.info("Preparing rename operations...")
	operations, longestSourceName := prepareRenameOperations(metadatas, cmdArgs.noPrefix)
	log.info(" done.\n")

	log.info("Verifying:\n")
	verifyOperations(operations, longestSourceName)
	log.info("done.\n")
	executeOperations(operations, cmdArgs.dryRun)
	log.info("\nFinished.\n")
}
