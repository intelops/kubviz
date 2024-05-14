package sdk

func (sdk *SDK) ConsumeNatsData(subject, consumerName string) error {
	data, err := sdk.natsClient.Consumer(subject, consumerName)
	if err != nil {
		return err
	}
	sdk.logger.Printf("Consumed successfully from stream %v", data)
	return nil
}
