package storage

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// ExportExpiredData exports expired data from a specific table in ClickHouse to an external storage (S3 in this case).
func ExportExpiredData(tableName string, db *sql.DB) error { // Create ClickHouse database client

	columns, err := getTableColumns(db, tableName)
	if err != nil {
		return fmt.Errorf("error getting columns for table %s: %v", tableName, err)
	}

	// Construct SQL query
	query := fmt.Sprintf("SELECT * FROM %s WHERE ExportedAt IS NULL", tableName)

	// Query expired data
	rows, err := db.Query(query)
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
		// for _, value := range columnValues {
		// 	rowData = append(rowData, fmt.Sprintf("%v", *value.(*interface{})))
		// }
		for _, value := range columnValues {
			// Dereference the pointer to get the interface{} value, then format it as a string
			rowData = append(rowData, fmt.Sprintf("%v", *value.(*interface{})))
		}
		csvline := strings.Join(rowData, ",") + "\n"
		_, err = csvFile.WriteString(fmt.Sprintf("%s\n", csvline))
		if err != nil {
			return fmt.Errorf("error writing into csv file: %v", err)
		}
	}

	// Update ExportedAt column with the current timestamp for exported rows
	updateQuery := fmt.Sprintf("ALTER TABLE %s UPDATE ExportedAt = now() WHERE ExportedAt IS NULL", tableName)
	_, err = db.Exec(updateQuery)
	if err != nil {
		return fmt.Errorf("error updating ExportedAt column: %v", err)
	}

	return nil
}

func getTableColumns(db *sql.DB, tableName string) (string, error) {
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

	return strings.Join(columns, ","), nil
}
