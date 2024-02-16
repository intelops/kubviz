package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/intelops/kubviz/client/pkg/config"
	"github.com/kelseyhightower/envconfig"
)

func ExportExpiredData(tableName string, db *sql.DB) error {
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

	// Construct CSV data in memory
	var csvData strings.Builder
	csvData.WriteString(columns + "\n") // Write CSV header

	for rows.Next() {
		// Assuming a dynamic structure, scan the columns into a slice of interface{}
		columnValues := make([]interface{}, len(strings.Split(columns, ",")))
		for i := range columnValues {
			columnValues[i] = new(interface{})
		}

		err := rows.Scan(columnValues...)
		if err != nil {
			return fmt.Errorf("error scanning ClickHouse row: %v", err)
		}

		// Write the values to the CSV data
		var rowData []string
		for _, value := range columnValues {
			// Dereference the pointer to get the interface{} value, then format it as a string
			rowData = append(rowData, fmt.Sprintf("%v", *value.(*interface{})))
		}
		csvData.WriteString(strings.Join(rowData, ",") + "\n")
	}

	// Upload the CSV data to S3
	err = uploadToS3(&csvData, fmt.Sprintf("exported_data_%s.csv", tableName))
	if err != nil {
		return fmt.Errorf("error uploading CSV to S3: %v", err)
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

func uploadToS3(csvData *strings.Builder, s3ObjectKey string) error {
	cfg := &config.Config{}
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatalf("Could not parse env Config: %v", err)
	}

	// Set up AWS S3 session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.AWSRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AWSAccessKey, cfg.AWSSecretKey, ""),
	})
	if err != nil {
		return fmt.Errorf("error creating S3 session: %v", err)
	}

	// Create an S3 service client
	s3Client := s3.New(sess)

	// Upload the CSV data to S3
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(cfg.S3BucketName),
		Key:    aws.String((s3ObjectKey)),
		Body:   strings.NewReader(csvData.String()),
	})
	if err != nil {
		return fmt.Errorf("error uploading data to S3: %v", err)
	}

	fmt.Printf("Data uploaded to S3: %s\n", s3ObjectKey)

	return nil
}
