package main

import (
	"context"
	"log"
	"time"

	"github.com/Glenn444/golang-chess/config"
	"github.com/Glenn444/golang-chess/internal/api"
	"github.com/Glenn444/golang-chess/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("error loading the config, ", err)
	}

	var DB_URL = config.DB_URL //os.Getenv("DB_URL")
	//var dbDriver = config.DBDriver
	var Address = config.ServerAddress

	//set up pool connection with config
	dbConfig, err := pgxpool.ParseConfig(DB_URL)
	if err != nil {
		log.Fatal("cannot parse db config: ", err)
	}
	
	dbConfig.MaxConns = 10
	dbConfig.MinConns = 2
	dbConfig.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		log.Fatal("cannot create db connection pool: ", err)
	}

	defer pool.Close()

	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	store := db.NewStore(pool)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("failed to set up server: ", err)
	}

	err = server.Start(Address)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
