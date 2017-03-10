package sparrow

import (
	"fmt"
	"github.com/wqx/sparrow/inject"
	"log"
	"net/http"
	"os"
	"reflect"
)

type Sparrow struct {
	inject.Injector
	handlers []Handler
	action   Handler
	logger   *log.Logger
}

// 工厂函数
func New() *Sparrow {
	s := &Sparrow{Injector: inject.New(), action: func() {}, logger: log.New(os.Stdout, "[sparrow] ", 0)}
	s.Map(s.logger)
	s.Map(defaultReturnHandler())
	return s
}

func (s *Sparrow) Handlers(handlers ...Handler) {
	s.handlers = make([]Handler, 0)
	for _, handler := range handlers {
		s.Register(handler)
	}
}

func (s *Sparrow) Action(handler Handler) {
	validateHandler(handler)
	s.action = handler
}

func (s *Sparrow) Logger(logger *log.Logger) {
	s.logger = logger
	s.Map(s.logger)
}

func (s *Sparrow) Register(handler Handler) {
	validateHandler(handler)

	s.handlers = append(s.handlers, handler)
}

func (s *Sparrow) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Dispatch
	s.createContext(w, r).run()
}

func (s *Sparrow) ListenAndServe(addr string) {
	logger := s.Injector.Get(reflect.TypeOf(s.logger)).Interface().(*log.Logger)

	logger.Printf("Listening on %s\n", addr)
	logger.Fatalln(http.ListenAndServe(addr, s))
}

func (s *Sparrow) createContext(w http.ResponseWriter, r *http.Request) *context {
	c := &context{inject.New(), s.handlers, s.action, NewResponseWriter(w), 0}
	c.SetParent(s)
	c.MapTo(c, (*Context)(nil))
	c.MapTo(c.rw, (*http.ResponseWriter)(nil))
	c.Map(r)
	return c
}

//////////////
type DefaultSparrow struct {
	*Sparrow
	Router
}

func Default() *DefaultSparrow {
	fmt.Printf("New Router...\n")
	r := NewRouter()
	fmt.Printf("New Sparrow...\n")
	s := New()
	fmt.Printf("Register Logger Handler...\n")
	s.Register(Logger())
	//...
	fmt.Printf("Map Routes...\n")
	s.MapTo(r, (*Routing)(nil))
	fmt.Printf("Add Post router...\n")
	s.Action(r.Handle)
	return &DefaultSparrow{s, r}
}
