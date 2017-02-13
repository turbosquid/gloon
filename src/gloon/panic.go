package main

import (
	"fmt"
	"log"
	"runtime/debug"
)

func handlePanic(r interface{}) {
	var msg string
	switch v := r.(type) {
	case error:
		msg = v.Error()
	case string:
		msg = v
	default:
		msg = fmt.Sprintf("PANIC (unknown type %#v)", v)
	}
	log.Printf("PANIC: %s", msg)
	log.Printf("Stack:\n%s", debug.Stack())
}
