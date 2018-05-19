package library

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"
)

// GetURL get url represent of host and port
func GetURL(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

// StringTags create tag string
func StringTags(tags ...interface{}) string {
	n := len(tags)
	if n == 0 {
		return ""
	}
	format := strings.Repeat("[%s]", n)
	return fmt.Sprintf(format, tags...)
}

// StackTrace returns the stack trace string on single line
func StackTrace() string {
	return strings.Replace(string(debug.Stack()), "\n", " <- ", -1)
}

// Recover will safely recover from any unexpected error panic
func Recover(next func(error)) {
	var err error
	r := recover()
	if r != nil {
		err = fmt.Errorf("unexpected %v", r)
		trace := StackTrace()
		log.Printf("%s >> trace: %s", err, trace)
	}
	if next != nil {
		next(err)
	}
}

// ToString converts anything to string
func ToString(any interface{}) string {
	switch any.(type) {
	case float32, float64:
		return fmt.Sprintf("%.6f", any)
	default:
		return fmt.Sprintf("%v", any)
	}
}

// JoinWithComma join strings with comma and returns it
func JoinWithComma(s []string) string {
	return strings.Join(s, ", ")
}
