package storage

import (
	"fmt"
	"os"

	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/client/pkg/config"
)

// ExportExpiredData exports expired data from a specific table in ClickHouse to an external storage (S3 in this case).
func ExportExpiredData(tableName string, conf *config.Config) error {
	// Create ClickHouse database client
	clickhouseDB, err := clickhouse.NewDBClient(conf)
	if err != nil {
		return fmt.Errorf("error creating ClickHouse database client: %v", err)
	}
	defer clickhouseDB.Close()

	columns, err := getTableColumns(clickhouseDB, tableName)
	if err != nil {
		return fmt.Errorf("error getting columns for table %s: %v", tableName, err)
	}

	// Construct SQL query
	query := fmt.Sprintf("SELECT * FROM %s WHERE ExportedAt IS NULL", tableName)

	// Query expired data
	rows, err := clickhouseDB.Query(query)
	if err != nil {
		return fmt.Errorf("error querying ClickHouse: %v", err)
	}
	defer rows.Close()

	// Create a CSV file to store the exported data
	csvFile, err := os.Create(fmt.Sprintf("exported_data_%s.csv", tableName))
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer csvFile.Close()

	// Write CSV header
	csvFile.WriteString(fmt.Sprintf("%s\n", columns))

	// Write rows to CSV
	for rows.Next() {
		// Assuming a dynamic structure, scan the columns into a slice of interface{}
		columnValues := make([]interface{}, len(columns))
		for i := range columnValues {
			columnValues[i] = new(interface{})
		}

		err := rows.Scan(columnValues...)
		if err != nil {
			return fmt.Errorf("error scanning ClickHouse row: %v", err)
		}

		// Write the values to the CSV file
		var rowData []string
		for _, value := range columnValues {
			rowData = append(rowData, fmt.Sprintf("%v", *value.(*interface{})))
		}

		csvFile.WriteString(fmt.Sprintf("%s\n", rowData))
	}

	// Update ExportedAt column with the current timestamp for exported rows
	updateQuery := fmt.Sprintf("ALTER TABLE %s UPDATE ExportedAt = now() WHERE ExportedAt IS NULL", tableName)
	_, err = clickhouseDB.Exec(updateQuery)
	if err != nil {
		return fmt.Errorf("error updating ExportedAt column: %v", err)
	}

	return nil
}

func getTableColumns(db clickhouse.DBInterface, tableName string) (string, error) {
	// Query to get column names
	query := fmt.Sprintf("DESCRIBE TABLE %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Get column names
	var columns []string
	for rows.Next() {
		var columnName string
		rows.Scan(&columnName)
		columns = append(columns, columnName)
	}

	return fmt.Sprintf("%s", columns), nil
}
