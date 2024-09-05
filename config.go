package main

import (
	"database/sql"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func InitRedis(name, port, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", name, port),
		Password: password, // no password set
		DB:       0,        // use default DB
	})
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
