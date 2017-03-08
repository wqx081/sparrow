package sparrow

import (
	"reflect"
)

type Handler interface{}

func validateHandler(handler Handler) {
	if reflect.TypeOf(handler).Kind() != reflect.Func {
		panic("sparrow handler must be a callable function")
	}
}
