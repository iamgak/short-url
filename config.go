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

func SeedData(db *sql.DB) error {
	// Create the url_shortner table if it doesn't exist
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS url_shortner (
            id INT AUTO_INCREMENT PRIMARY KEY,
            long_url VARCHAR(255) NOT NULL,
            hash VARCHAR(255) DEFAULT NULL,
			active TINYINT(1) DEFAULT 0,
			traffic INT DEFAULT 0,
			user_id INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP 
        )AUTO_INCREMENT = 1000;`)

	return err
}
