package sparrow

import (
	"net/http"
)

type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Hijacker

	Status() int
	Written() bool
	Size() int
	Before(BeforeFunc)
}

type BeforeFunc func(ResponseWriter)
