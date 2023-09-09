package main

import (
	"fmt"
	"log"
	"os"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/golang-migrate/migrate/v4"
	cm "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/intelops/kubviz/client/pkg/config"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "migration-cli",
	Short: "CLI for managing migrations",
	Long:  `A CLI tool developed to manage migrations for ClickHouse.`,
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run migrations for ClickHouse",
	Run: func(cmd *cobra.Command, args []string) {
		migrationsPath := os.Getenv("MIGRATIONS_PATH")

		cfg := &config.Config{}
		if err := envconfig.Process("", cfg); err != nil {
			log.Fatalf("Could not parse env Config: %v", err)
		}

		conn := clickhouse.OpenDB(&clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:%d", cfg.DBAddress, cfg.DbPort)},
		})
		if err := conn.Ping(); err != nil {
			if exception, ok := err.(*clickhouse.Exception); ok {
				fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
			} else {
				fmt.Println(err)
			}
			log.Fatalf("Could not ping the DB connection: %v", err)
		}

		driver, err := cm.WithInstance(conn, &cm.Config{})
		if err != nil {
			log.Fatalf("Failed to create migrate driver: %v", err)
		}

		m, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s", migrationsPath),
			"clickhouse",
			driver,
		)
		if err != nil {
			log.Fatalf("Migration initialization failed: %v", err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration failed: %v", err)
		}

		fmt.Println("Migrations completed successfully!")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
