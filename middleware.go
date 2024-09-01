package main

import (
	"net/http"
)

func (app *application) middleware(next http.Handler) http.Handler {
	// next
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
