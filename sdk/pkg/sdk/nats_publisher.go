package sdk

func (sdk *SDK) PublishToNats(subject string, streamName string, data interface{}) error {
	if err := sdk.natsClient.Publish(subject, streamName, data); err != nil {
		return err
	}
	sdk.logger.Printf("Message published successfully to stream %v", streamName)
	return nil
}
