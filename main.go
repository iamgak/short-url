package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql" // sql pool register
	"github.com/joho/godotenv"
)

type application struct {
	Infolog  *log.Logger
	Errorlog *log.Logger
	Shortner *ShortnerModel
	User_id  int
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	RedisName := os.Getenv("R_NAME")
	RedisPassword := os.Getenv("R_PASSW")
	RedisPort := os.Getenv("R_PORT")
	dsn := flag.String("dsn", fmt.Sprintf("%s:@/%s?parseTime=true", dbUser, dbName), "MySQL data source name")
	// dsn := flag.String("dsn", fmt.Sprintf("%s:@/%s?parseTime=true", dbUser, dbName), "MySQL data source name")
	flag.Parse()
	sql, err := openDB(*dsn)
	if err != nil {
		panic(err)
	}

	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	logger := log.New(os.Stdout, "URL-Shortner ", log.Ldate|log.Lshortfile)
	client := InitRedis(RedisName, RedisPort, RedisPassword)
	app := application{
		Infolog:  logger,
		Errorlog: errorLog,
		Shortner: Init(sql, client),
	}
	port := flag.String("port", ":8080", "Http Connection Port Addres")

	serve := &http.Server{
		Addr:    *port,
		Handler: app.routes(),
	}

	app.Infolog.Print("URL-Shortner is Ready to cut your URL. Till the date of Judgement(365billion records) !!")
	err = serve.ListenAndServe()
	if err != nil {
		panic(err)
	}

}
