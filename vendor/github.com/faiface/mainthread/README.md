# mainthread [![GoDoc](https://godoc.org/github.com/faiface/mainthread?status.svg)](http://godoc.org/github.com/faiface/mainthread) [![Report card](https://goreportcard.com/badge/github.com/faiface/mainthread)](https://goreportcard.com/report/github.com/faiface/mainthread)

Package mainthread allows you to run code on the main operating system thread.

`go get github.com/faiface/mainthread`

Operating systems often require, that code which deals with windows and graphics has to run on the
main thread. This is however somehow challenging in Go due to Go's concurrent nature.

This package makes it easily possible.

All you need to do is put your main code into a separate function and call `mainthread.Run` from
your real main, like this:

```go
package main

import (
	"fmt"

	"github.com/faiface/mainthread"
)

func run() {
	// now we can run stuff on the main thread like this
	mainthread.Call(func() {
		fmt.Println("printing from the main thread")
	})
	fmt.Println("printing from another thread")
}

func main() {
	mainthread.Run(run) // enables mainthread package and runs run in a separate goroutine
}
```

## More functions

If you don't wish to wait until a function finishes running on the main thread, use
`mainthread.CallNonBlock`:

```go
mainthread.CallNonBlock(func() {
	fmt.Println("i'm in the main thread")
})
fmt.Println("but imma be likely printed first, cuz i don't wait")
```

If you want to get some value returned from the main thread, you can use `mainthread.CallErr` or
`mainthread.CallVal`:

```go
err := mainthread.CallErr(func() error {
	return nil // i don't do nothing wrong
})
val := mainthread.CallVal(func() interface{} {
	return 42 // the meaning of life, universe and everything
})
```

If `mainthread.CallErr` or `mainthread.CallVal` aren't sufficient for you, you can just assign
variables from within the main thread:

```go
var x, y int
mainthread.Call(func() {
	x, y = 1, 2
})
```

However, be careful with `mainthread.CallNonBlock` when dealing with local variables.
