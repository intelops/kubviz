package clickhouse

import (
	"context"
	"errors"
	"strings"
	"time"
)

func (c *Client) InsertData(tableName string, data interface{}) error {
	ctx := context.Background()

	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors.New("data is not in the expected format")
	}

	columns := make([]string, 0, len(dataMap))
	values := make([]interface{}, 0, len(dataMap))
	placeholders := make([]string, 0, len(dataMap))

	for column, value := range dataMap {
		columns = append(columns, column)
		values = append(values, value)
		placeholders = append(placeholders, "?")
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO "+tableName+" ("+strings.Join(columns, ",")+") VALUES ("+strings.Join(placeholders, ",")+")")
	if err != nil {
		return err
	}
	defer stmt.Close()

	values = append(values, time.Now().UTC())

	_, err = stmt.ExecContext(ctx, values...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *Client) List(input interface{}) ([]map[string]interface{}, error) {
	var dataList []map[string]interface{}

	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return nil, errors.New("input is not a map[string]interface{}")
	}

	var traverse func(m map[string]interface{})
	traverse = func(m map[string]interface{}) {
		dataList = append(dataList, m)

		for _, v := range m {
			if subMap, ok := v.(map[string]interface{}); ok {
				traverse(subMap)
			}
		}
	}

	traverse(inputMap)

	return dataList, nil
}
