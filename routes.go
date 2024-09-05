package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	auth := alice.New(app.LoginMiddleware)

	//home related routes
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/get/:hash", app.redirect)
	router.Handler(http.MethodPost, "/create/shortner", auth.ThenFunc(app.add_url))
	// router.HandleFunc("/add/custom_url", app.custom_add_url)
	return secureHeaders(app.logRequest(router))
}
