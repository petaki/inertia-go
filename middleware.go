package inertia

import "net/http"

// Middleware function.
func (i *Inertia) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(HeaderInertia) == "" {
			next.ServeHTTP(w, r)

			return
		}

		if r.Method == "GET" && r.Header.Get(HeaderVersion) != i.version {
			w.Header().Set(HeaderLocation, i.url+r.RequestURI)
			w.WriteHeader(http.StatusConflict)

			return
		}

		next.ServeHTTP(w, r)
	})
}
