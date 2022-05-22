package main

import (
	"log"
	"net/http"

	"github.com/ivolff/graphql-mafia/db"
	"github.com/ivolff/graphql-mafia/graph"
	"github.com/ivolff/graphql-mafia/graph/generated"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const (
	DB_PORT        = 27017
	DB_HOST        = "mongodb"
	DB_USERNAME    = "root"
	DB_PASS        = "example"
	QUERY_ENDPOINT = "/query"
	PORT           = "8080"
)

func main() {
	dbHandle := db.MongoDbHandle{}
	err := dbHandle.InitConnection(DB_USERNAME, DB_PASS, DB_HOST, DB_PORT)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to mongoDB", err)
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{DbHandle: dbHandle}}))

	http.Handle("/", playground.Handler("Mafia GraphQL", QUERY_ENDPOINT))
	http.Handle(QUERY_ENDPOINT, srv)

	log.Printf("serving on http://localhost:%s/", PORT)
	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}
