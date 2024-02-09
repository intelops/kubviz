package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/client/pkg/config"
	"github.com/intelops/kubviz/graphqlserver/graph"
	"github.com/kelseyhightower/envconfig"
)

const defaultPort = "8085"
const (
	maxRetries = 5
	retryDelay = 5 * time.Second
)

func main() {
	log.Println("Graph ql server starting ... Iteration one")
	cfg := &config.Config{}
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatalf("Could not parse env Config: %v", err)
	}
	db, err := initializeDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	resolver := graph.NewResolver(db)
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func initializeDatabase(cfg *config.Config) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		_, db, err = clickhouse.NewDBClient(cfg)
		if err == nil {
			log.Println("Successfully connected to the database")
			return db, nil
		}
		log.Printf("Failed to connect to database, retrying (%d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	// If the loop exits and the connection is not established, return the error
	return nil, err
}
