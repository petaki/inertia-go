package inertia

import "net/http"

type responseWriter struct {
	http.ResponseWriter
}

func (rw *responseWriter) WriteHeader(code int) {
	if code == http.StatusFound {
		code = http.StatusSeeOther
	}

	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}
