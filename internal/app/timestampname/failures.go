// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

import (
	"errors"
	"fmt"
	"strings"
)

func _raise(file string, descriptor string, err error) {
	var sb strings.Builder
	sb.WriteString("Failure:")
	if len(file) > 0 {
		sb.WriteString("\n\tFile:       ")
		sb.WriteString(file)
	}
	sb.WriteString("\n\tDescriptor: ")
	sb.WriteString(descriptor)
	if err != nil {
		sb.WriteString("\n\tError:      ")
		sb.WriteString(err.Error())
	}
	panic(errors.New(sb.String()))
}

func Raise(file string, descriptor string) {
	_raise(file, descriptor, nil)
}

func RaiseErr(file string, descriptor string, err error) {
	_raise(file, descriptor, err)
}

func RaiseFmt(format string, a ...interface{}) {
	_raise("", fmt.Sprintf(format, a...), nil)
}

func RaiseFmtFile(file string, format string, a ...interface{}) {
	_raise(file, fmt.Sprintf(format, a...), nil)
}

func Catch(err error, description string) {
	if err != nil {
		RaiseErr("", description, err)
	}
}

func CatchFile(err error, file string, descriptor string) {
	if err != nil {
		RaiseErr(file, descriptor, err)
	}
}
