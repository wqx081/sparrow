package main

import (
	"github.com/wqx/sparrow"
)

func main() {
	sparrow := sparrow.Default()
	sparrow.GET("/", func() string { return "Hello world"})
	sparrow.ListenAndServe(":9090")
}
