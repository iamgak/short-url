package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/julienschmidt/httprouter"
	"github.com/redis/go-redis/v9"
)

type Message struct {
	Status  bool
	Message any
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.ErrorMessage(w, http.StatusNotFound, "Error404 Page Not found")
		return
	}

	fmt.Fprint(w, "Hello From URL Shortner !!")
}

func (app *application) redirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		app.ErrorMessage(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	params := httprouter.ParamsFromContext(r.Context())
	hash := params.ByName("hash")

	pattern := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !pattern.MatchString(hash) {
		app.ErrorMessage(w, http.StatusNotFound, "Error404 Page Not found")
		return
	}

	val, err := app.Shortner.RedisGet(hash)
	if err == nil {
		app.FinalMessage(w, 200, fmt.Sprintf("your URL is %s", val))
		return
	} else if err != sql.ErrNoRows && err != redis.Nil {
		app.ErrorMessage(w, http.StatusInternalServerError, "Internal Server Error")
		app.Errorlog.Print(err)
		return
	}

	url, err := app.Shortner.GetShortner(hash)
	if err != nil {
		if err == sql.ErrNoRows {
			app.ErrorMessage(w, http.StatusInternalServerError, "Error404 Page Not found")
			return
		}
	}

	err = app.Shortner.IncrementHit(hash)
	if err != nil {
		// Internal Server Error
		app.ErrorMessage(w, http.StatusInternalServerError, "Internal Server Error")
		app.Errorlog.Print(err)
		return
	}

	err = app.Shortner.RedisSet(hash, url)
	if err != nil {
		app.ErrorMessage(w, http.StatusInternalServerError, "Internal Server Error")
		app.Errorlog.Print(err)
		return
	}

	http.Redirect(w, r, url, http.StatusPermanentRedirect)
}

func (app *application) add_url(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		app.ErrorMessage(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.ErrorMessage(w, http.StatusInternalServerError, "Internal Server Error")
		app.Errorlog.Print(err)
		return
	}

	long_url := r.PostForm.Get("long_url")

	if long_url == "" {
		app.ErrorMessage(w, http.StatusNotFound, "URL Not Found")
		return
	}

	hash, err := app.Shortner.CreateShortner(long_url, app.User_id)
	if err != nil && err != sql.ErrNoRows {
		app.ErrorMessage(w, http.StatusInternalServerError, "Internal Server Error")
		app.Errorlog.Print(err)
		return
	}

	err = app.Shortner.RedisSet(hash, long_url)
	if err != nil {
		app.ErrorMessage(w, http.StatusInternalServerError, "Internal Server Error")
		app.Errorlog.Print(err)
		return
	}

	fmt.Printf("Your Short Url is created %s", hash)
}

func (app *application) ErrorMessage(w http.ResponseWriter, statusCode int, Message any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(struct {
		Status bool
		Error  any
	}{
		Status: false,
		Error:  Message,
	})
}

func (app *application) FinalMessage(w http.ResponseWriter, statusCode int, Message any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(struct {
		Success bool
		Message any
	}{
		Success: true,
		Message: Message,
	})
}
