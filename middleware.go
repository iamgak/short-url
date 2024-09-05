package main

import (
	"net/http"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.Infolog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *application) LoginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("ldata")
		if err != nil || cookie.Value == "" {
			if err != nil {
				if err == http.ErrNoCookie {
					app.Errorlog.Print("Cookie not found")
				} else {
					app.Errorlog.Print(err)
				}
			}

			app.ErrorMessage(w, http.StatusUnauthorized, "User need to Login")
			return
		}

		app.Infolog.Print(cookie.Value)
		userID, err := app.Shortner.RedisGet(cookie.Value)
		if err != nil {
			if err != redis.Nil {
				app.Errorlog.Print(err)
			}

			app.ErrorMessage(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		app.User_id, err = strconv.Atoi(userID)
		if err != nil {
			app.ErrorMessage(w, http.StatusInternalServerError, "Internal Server Error")
			app.Errorlog.Print(err)
			return
		}

		if app.User_id == 0 {
			app.ErrorMessage(w, http.StatusUnauthorized, "User need to Login")
			return
		}

		next.ServeHTTP(w, r)

	})
}
