package sparrow

import (
	"log"
	"net/http"
	"time"
)

func Logger() Handler {
	return func(w http.ResponseWriter, r *http.Request, c Context, log *log.Logger) {
		start := time.Now()

		addr := r.Header.Get("X-Real-IP")
		if addr == "" {
			addr = r.Header.Get("X-Forwarded-For")
			if addr == "" {
				addr = r.RemoteAddr
			}
		}

		log.Printf("Started %s %s for %s", r.Method, r.URL.Path, addr)
		rw := w.(ResponseWriter)
		c.Next()

		log.Printf("Completed %v %s in %v\n", rw.Status(), http.StatusText(rw.Status()), time.Since(start))
	}
}
