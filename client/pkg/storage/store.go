package storage

import (
	"fmt"
	"os"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/client/pkg/config"
)

// const (
// 	clickhouseDSN = "tcp://clickhouse-server:9000?username=default&password=&database=your_database"
// 	s3Bucket      = "your-s3-bucket"
// 	s3ObjectKey   = "exported_data.csv"
// )

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

	query := fmt.Sprintf("SELECT * FROM %s WHERE ExpiryDate <= now() + toIntervalDay(1)", tableName)
	// oneDayBefore := time.Now().Add(-24 * time.Hour).Format("2006-01-02")

	// // Construct SQL query to get data one day before the expiry date
	// query := fmt.Sprintf("SELECT * FROM %s WHERE ExpiryDate <= '%s'", tableName, oneDayBefore)

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

	// Upload the CSV file to S3
	// err = uploadToS3(fmt.Sprintf("exported_data_%s.csv", tableName))
	// if err != nil {
	// 	return fmt.Errorf("error uploading to S3: %v", err)
	// }

	// fmt.Printf("Data exported and uploaded to S3 successfully for table %s.\n", tableName)
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

// func uploadToS3(filename string) error {
// 	sess, err := session.NewSession(&aws.Config{
// 		Region: aws.String("your-region"),
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	file, err := os.Open(filename)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	s3Client := s3.New(sess)
// 	_, err = s3Client.PutObject(&s3.PutObjectInput{
// 		Bucket: aws.String(s3Bucket),
// 		Key:    aws.String(s3ObjectKey),
// 		Body:   file,
// 	})
// 	if err != nil {
// 		if awsErr, ok := err.(awserr.Error); ok {
// 			fmt.Println("AWS Error:", awsErr.Code(), awsErr.Message())
// 		}
// 		return err
// 	}

// 	return nil
// }
