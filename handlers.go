package main

import (
	"fmt"
	"log"
	"net/http"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fmt.Fprint(w, "Hello From URL Shortner !!")
}

func (app *application) get_url(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Use the r.PostForm.Get() method to retrieve the title and content
	// from the r.PostForm map.
	url := r.PostForm.Get("url")
	app.Infolog.Print(url)
	fmt.Fprint(w, "Hello From URL Shortner Worker  !!", app.get_shortner(url))
}

func (app *application) custom_add_url(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Use the r.PostForm.Get() method to retrieve the title and content
	// from the r.PostForm map.
	url := r.PostForm.Get("url")
	fmt.Fprint(w, "Hello From URL Shortner Worker  !!", app.create_custom_shortner(url))
}

func (app *application) add_url(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	var id int64 = 1
	log.Print(app.create_shortner(id))
	fmt.Fprint(w, "Hello From URL Shortner Worker  !!", app.create_shortner(id))
}
