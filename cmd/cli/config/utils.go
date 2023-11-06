package config

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/golang-migrate/migrate/v4"
	ch "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (cfg *Config) openClickHouseConn() (*sql.DB, error) {

	var options clickhouse.Options
	if cfg.ClickHouseUsername != "" && cfg.ClickHousePassword != "" {
		fmt.Println("Using provided username and password")
		options = clickhouse.Options{
			Addr:  []string{fmt.Sprintf("%s:%d", cfg.DBAddress, cfg.DbPort)},
			Debug: true,
			Auth: clickhouse.Auth{
				Username: cfg.ClickHouseUsername,
				Password: cfg.ClickHousePassword,
			},
		}

	} else {
		fmt.Println("Using connection without username and password")
		options = clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:%d", cfg.DBAddress, cfg.DbPort)},
		}
	}

	conn := clickhouse.OpenDB(&options)
	if err := conn.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			return nil, fmt.Errorf("[%d] %s %s", exception.Code, exception.Message, exception.StackTrace)
		} else {
			return nil, err
		}
	}
	return conn, nil
}

func (cfg *Config) processSQLTemplate(filePath string) (string, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	tmpl, err := template.New("sql").Parse(string(data))
	if err != nil {
		return "", err
	}

	params := map[string]string{
		"TTLValue": cfg.TtlInterval,
		"TTLUnit":  cfg.TtlUnit,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, params)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (cfg *Config) Migrate() error {
	dir := cfg.SchemaPath
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			fullPath := filepath.Join(dir, file.Name())
			processedSQL, err := cfg.processSQLTemplate(fullPath)
			if err != nil {
				return fmt.Errorf("failed to process the sql template for file %s : %w", file.Name(), err)
			}

			err = os.WriteFile(fullPath, []byte(processedSQL), 0644)
			if err != nil {
				return fmt.Errorf("failed to write to file %s : %w", fullPath, err)
			}
		}
	}

	conn, err := cfg.openClickHouseConn()
	if err != nil {
		return fmt.Errorf("unable to create a clickhouse conection %w", err)
	}

	driver, err := ch.WithInstance(conn, &ch.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", dir),
		"clickhouse",
		driver,
	)
	if err != nil {
		return fmt.Errorf("clickhouse migration initialization failed %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed %w", err)
	}
	fmt.Println("Clickhouse Migration applied successfully!")
	return nil
}
