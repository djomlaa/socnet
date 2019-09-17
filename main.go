package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/djomlaa/socnet/internal/handler"
	"github.com/djomlaa/socnet/internal/service"
	"github.com/hako/branca"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	dbport     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
	schema   = "socnet"
	port     = 8789
)

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable search_path=%s",
	host, dbport, user, password, dbname, schema)
	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatalf("Could not open db connection: %v\n", err)
		return
	}

	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Could not ping to db: %v\n", err)
		return
	}

	// TODO: use service.TokenLifespan with branca
	codec := branca.NewBranca("supersecretkeyyoushouldnotcommit")
	codec.SetTTL(uint32(service.TokenLifespan.Seconds()))

	s := service.New(db, codec)

	h := handler.New(s)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("accepting connections on port %d", port)

	if err = http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("could not start server: %v\n", err)
	}
}
