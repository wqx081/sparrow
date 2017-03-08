package sparrow

import (
	"github.com/wqx/sparrow/inject"
)

// 请求上下文
type Context interface {
	inject.Injector
	Next()
	Written() bool
}

type context struct {
	inject.Injector
	handlers []Handler
	post     Handler
	rw       ResponseWriter
	index    int
}

func (c *context) handler() Handler {
	if c.index < len(c.handlers) {
		return c.handlers[c.index]
	}
	if c.index == len(c.handlers) {
		return c.post
	}
	panic("invalid index for context handler")
}

func (c *context) Next() {
	c.index += 1
	c.run()
}

func (c *context) Written() bool {
	return c.rw.Written()
}

func (c *context) run() {
	for c.index <= len(c.handlers) {
		// 获取对应的 handler 并且执行
		_, err := c.Invoke(c.handler())
		if err != nil {
			panic(err)
		}
		c.index += 1

		if c.Written() {
			return
		}
	}
}
