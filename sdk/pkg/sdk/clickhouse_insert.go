package sdk

func (sdk *SDK) ClickHouseInsertData(tableName string, data interface{}) error {
	err := sdk.clickhouseClient.InsertData(tableName, data)
	if err != nil {
		return err
	}
	sdk.logger.Printf("insert into table successfully %v", data)
	return nil
}
