// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

//
// BEFORE INITIALIZATION
//

type commandLineArguments struct {
	dryRun      bool
	noPrefix    bool
	debugOutput bool
	timezone    *time.Location
}

func parseCommandLineArguments() commandLineArguments {
	var cmdArgs commandLineArguments
	flag.BoolVar(&cmdArgs.dryRun, "dry", false, "dry run")
	flag.BoolVar(&cmdArgs.noPrefix, "noprefix", false, "no counter prefix")
	flag.BoolVar(&cmdArgs.debugOutput, "debug", false, "debug output")
	var zoneOffsetString string
	flag.StringVar(&zoneOffsetString, "timezone", "0", "time zone where the video was taken. May be signed, single digit or 4 digits.")
	flag.Parse()

	// parsing zone offset:
	switch len(zoneOffsetString) {
	case 1:
		zoneOffsetString = "+" + zoneOffsetString
		fallthrough
	case 2:
		if zoneOffsetString[0] != '-' && zoneOffsetString[0] != '+' {
			RaiseFmt("invalid time zone offset: %s", zoneOffsetString)
		}
		hours, err := strconv.Atoi(zoneOffsetString)
		Catch(err, "invalid time zone offset")
		cmdArgs.timezone = time.FixedZone("UTC"+zoneOffsetString, hours*60*60)
	case 4:
		hours, err := strconv.Atoi(zoneOffsetString[0:2])
		Catch(err, "invalid time zone offset")
		minutes, err := strconv.Atoi(zoneOffsetString[2:])
		cmdArgs.timezone = time.FixedZone("UTC+"+zoneOffsetString, hours*60*60+minutes*60)
	case 5:
		if zoneOffsetString[0] != '-' && zoneOffsetString[0] != '+' {
			RaiseFmt("invalid time zone offset: %s", zoneOffsetString)
		}
		hours, err := strconv.Atoi(zoneOffsetString[0:3])
		Catch(err, "invalid time zone offset")
		minutes, err := strconv.Atoi(zoneOffsetString[3:])
		if hours < 0 {
			minutes = -minutes
		}
		cmdArgs.timezone = time.FixedZone("UTC+"+zoneOffsetString, hours*60*60+minutes*60)
	default:
		RaiseFmt("invalid time zone offset: %s", zoneOffsetString)
	}

	debug("command line arguments: %v", cmdArgs)
	return cmdArgs
}

//
// END BEFORE INITIALIZATION
//

var (
	cmdArgs commandLineArguments
	workDir string
)

//
// LOGGING
//

func debug(format string, a ...interface{}) {
	if cmdArgs.debugOutput {
		fmt.Fprintf(os.Stdout, "\033[32m"+format+"\033[0m\n", a...)
	}
}

func info(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format, a...)
}

//
// END LOGGING
//

func processFiles(files []inputFile) []fileMetadata {
	var total = len(files)
	var output = make([]fileMetadata, total)
	for index, file := range files {
		info("\rProcessing files: %d/%d...", index+1, total)
		output[index] = fileMetadataCreationTimestamp(file)
	}
	info(" done.\n")
	return output
}

func verifyOperations(operations []renameOperation, longestSourceName int) {
	duplicatesMap := make(map[string]string)
	for _, operation := range operations {
		info("    %[3]*[1]s    =>    %[2]s\n", operation.from, operation.to, longestSourceName)
		// check for target name duplicates:
		if _, existsInMap := duplicatesMap[operation.to]; existsInMap {
			Raise(operation.to, "duplicate rename")
		} else {
			duplicatesMap[operation.to] = operation.to
		}
		// check for renaming duplicates:
		if operation.from != operation.to {
			if _, existsInDir := os.Stat(operation.to); existsInDir == nil {
				Raise(operation.to, "exists on file system")
			}
		}
	}
}

func executeOperations(operations []renameOperation, dryRun bool) {
	for index, operation := range operations {
		info("\rRenaming files: %d/%d", index+1, len(operations))
		if !dryRun {
			renameErr := os.Rename(operation.from, operation.to)
			CatchFile(renameErr, operation.from, "rename")
			chmodErr := os.Chmod(operation.to, 0444)
			CatchFile(chmodErr, operation.from, "chmod")
		}
	}
	info(" done.\n")
}

func Exec() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "\n\033[31m%v\033[0m\n", r)
		}
	}()

	cmdArgs = parseCommandLineArguments()

	info("Scanning for files... ")
	var err error
	workDir, err = os.Getwd()
	Catch(err, "failed to get current working directory")
	var inputFiles = listFiles(workDir)
	info("%d supported files found.\n", len(inputFiles))

	metadatas := processFiles(inputFiles)
	info("Preparing rename operations...")
	operations, longestSourceName := prepareRenameOperations(metadatas, cmdArgs.noPrefix)
	info(" done.\n")

	info("Verifying:\n")
	verifyOperations(operations, longestSourceName)
	info("done.\n")
	executeOperations(operations, cmdArgs.dryRun)
	info("\nFinished.\n")
}
