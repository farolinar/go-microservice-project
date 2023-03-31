package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/farolinar/go-microservice-authentication/data"
)

var webPort = "80"
var counts int64

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Printf("Starting authentication service at port %s\n", webPort)

	// connect to database
	conn := connectToDB()
	if conn == nil {
		log.Panic("can't connect to postgres...")
	}

	// set up config
	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {

		return nil, err
	}

	err = db.Ping()
	if err != nil {

		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("postgres not yet ready...")
			counts++
		} else {
			log.Println("connected to postgres!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
		}

		log.Println("backing off for two seconds...")
		time.Sleep(2 * time.Second)
		continue
	}
}
