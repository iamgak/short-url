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
	router.HandlerFunc(http.MethodGet, "/get/:shortner", app.redirect)
	router.Handler(http.MethodPost, "/create/shortner", auth.ThenFunc(app.add_url))
	router.Handler(http.MethodPost, "/disable/:shortner", auth.ThenFunc(app.remove_url))
	// router.Handler(http.MethodGet, "/api/dbg/", expvar.Handler())

	standard := alice.New(app.rateLimit, secureHeaders, app.recoverPanic, app.metrics)
	return standard.Then(router)
}
