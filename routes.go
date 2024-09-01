package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/get/shortner", app.get_url)
	mux.HandleFunc("/create/shortner", app.add_url)
	mux.HandleFunc("/add/custom_url", app.custom_add_url)
	return app.middleware(mux)
}
