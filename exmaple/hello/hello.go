package main

import (
	_ "fmt"
	"github.com/wqx/sparrow"
)

func main() {
	sparrow := sparrow.New()
	sparrow.ListenAndServe(":9090")
}
