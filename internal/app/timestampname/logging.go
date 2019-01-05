// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"fmt"
	"os"
)

type logger interface {
	fatalityCheck(err error, format string, a ...interface{})
	fatalityDo(format string, a ...interface{})
	info(format string, a ...interface{})
	debug(format string, a ...interface{})
}

type consoleLogger struct {
	debugEnabled bool
}

func newLog(debug bool) logger {
	l := new(consoleLogger)
	l.debugEnabled = debug
	return l
}

func (log consoleLogger) fatalityDo(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "\n\033[31m"+format+"\033[0m\n", a...)
	os.Exit(1)
}

func (log consoleLogger) fatalityCheck(err error, format string, a ...interface{}) {
	if err != nil {
		log.fatalityDo(format, a...)
	}
}

func (log consoleLogger) info(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format, a...)
}

func (log consoleLogger) debug(format string, a ...interface{}) {
	if log.debugEnabled {
		fmt.Fprintf(os.Stdout, "\033[32m"+format+"\033[0m", a...)
	}
}
