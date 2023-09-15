package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	cm "github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/intelops/kubviz/cmd/cli/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "migration",
	Short: "CLI for managing migrations",
	Long:  `A CLI tool developed to manage migrations for Kubviz Client ClickHouse.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var sqlCmd = &cobra.Command{
	Use:   "sql",
	Short: "Manage SQL migrations",
	Long: `The sql subcommand is used to manage SQL migrations for Kubviz Client ClickHouse.
You can execute migrations using the -e flag and confirm with --yes.`,
	Example: `
# Execute migrations and confirm
migration sql -e --yes

# Execute migrations without confirmation
migration sql -e --no`,
	Run: func(cmd *cobra.Command, args []string) {
		executeFlag, _ := cmd.Flags().GetBool("execute")
		yesFlag, _ := cmd.Flags().GetBool("yes")

		if !executeFlag && !yesFlag {
			cmd.Help()
			return
		}

		if executeFlag {
			if yesFlag {
				db, cfg, err := config.OpenClickHouseConn()
				if err != nil {
					log.Fatalf("Failed to open ClickHouse connection: %v", err)
				}
				defer db.Close()
				driver, err := cm.WithInstance(db, &cm.Config{})
				if err != nil {
					log.Fatalf("Failed to create migrate driver: %v", err)
				}

				m, err := migrate.NewWithDatabaseInstance(
					fmt.Sprintf("file://%s", cfg.SchemaPath),
					"clickhouse",
					driver,
				)
				if err != nil {
					log.Fatalf("Clickhouse Migration initialization failed: %v", err)
				}
				if err := m.Up(); err != nil && err != migrate.ErrNoChange {
					log.Fatalf("Migration failed: %v", err)
				}
				fmt.Println("Clickhouse Migrations applied successfully!")
			} else {
				fmt.Println("Clickhouse Migration skipped due to --no flag.")
			}
		}
	},
}

func init() {
	sqlCmd.Flags().BoolP("execute", "e", false, "Execute the migrations")
	sqlCmd.Flags().BoolP("yes", "y", false, "Confirm execution")

	rootCmd.AddCommand(sqlCmd)
}
