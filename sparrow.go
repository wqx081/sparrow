package sparrow

import (
	"io"
	"log"
	"net/http"
)

type Sparrow struct {
}

// 工厂函数
func New() *Sparrow {
	return &Sparrow{}
}

func (s *Sparrow) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Dispatch
	io.WriteString(w, "Hello, world!")
}

func (s *Sparrow) ListenAndServe(addr string) {
	log.Printf("Listening on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, s))
}
