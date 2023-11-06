package commands

import (
	"fmt"
	"log"
	"os"

	_ "github.com/golang-migrate/migrate/v4/source/file"
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
				cfg, err := config.New()
				if err != nil {
					log.Fatalf("failed to parse the env : %v", err.Error())
					return
				}
				if err := cfg.Migrate(); err != nil {
					log.Fatalf("failed to migrate : %v", err.Error())
					return
				}
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
