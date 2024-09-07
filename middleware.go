package main

import (
	"expvar"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
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

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic
		// as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a panic or
			// not.
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after a response has been
				// sent.
				w.Header().Set("Connection", "close")
				// The value returned by recover() has the type interface{}, so we use
				// fmt.Errorf() to normalize it into an error and call our
				// serverErrorResponse() helper. In turn, this will log the error using
				// our custom Logger type at the ERROR level and send the client a 500
				// Internal Server Error response.
				app.ErrorMessage(w, 404, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Declare a mutex and a map to hold the clients' IP addresses and rate limiters.
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// this part is to release older ip
	go func() {
		for {
			// it will run until code run but take break every minute laziness
			time.Sleep(time.Minute)
			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place.
			mu.Lock()
			// Loop through all clients. If they haven't been seen within the last three
			// minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// to print Remote Addr, Protocol, Method, URL path in log
		app.Infolog.Printf("ip_addr-%s  http-%s method-%s uri-%s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.ErrorMessage(w, 501, "Internal Server Error")
			app.Errorlog.Print(err)
			return
		}

		// Lock the mutex to prevent this code from being executed concurrently.
		mu.Lock()

		if _, found := clients[ip]; !found {
			// Create and add a new client struct to the map if it doesn't already exist.
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Limit(app.Limiter.rps), app.Limiter.burst),
			}
		}

		//main work is done here concurrently checking all the clients[ip] if there is more than given burst, rps than it wont allow
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.ErrorMessage(w, http.StatusTooManyRequests, "Too, many request. Rate Limit Exceed")
			return
		}

		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func (app *application) metrics(next http.Handler) http.Handler {
	// Initialize the new expvar variables when the middleware chain is first built.
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_μs")
	// The following code will be run for every request...
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record the time that we started to process the request.
		start := time.Now()
		// Use the Add() method to increment the number of requests received by 1.
		totalRequestsReceived.Add(1)
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
		// On the way back up the middleware chain, increment the number of responses
		// sent by 1.
		totalResponsesSent.Add(1)
		// Calculate the number of microseconds since we began to process the request,
		// then increment the total processing time by this amount.
		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}
