package chromedptest

import (
	"fmt"
)

type LogFunc func(format string, a ...interface{})

var log = LogFunc(nil)

func SetLogger(lf LogFunc) {
	log = lf
}
func Printf(format string, a ...interface{}) {
	if log == nil {
		return
	}

	fmt.Printf(format, a...)
}
