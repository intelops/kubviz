package sdk

func (sdk *SDK) ListtData(data interface{}) error {
	data, err := sdk.clickhouseClient.List(data)
	if err != nil {
		return err
	}
	sdk.logger.Printf("insert into table successfully %v", data)
	return nil
}
