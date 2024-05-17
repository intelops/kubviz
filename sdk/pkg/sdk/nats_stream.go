package sdk

func (sdk *SDK) CreateNatsStream(streamName string, streamSubjects []string) error {
	if err := sdk.natsClient.CreateStream(streamName, streamSubjects); err != nil {
		return err
	}
	sdk.logger.Printf("Stream created successfully for streamName %v, streamSubjects %v", streamName, streamSubjects)
	return nil
}
