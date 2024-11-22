package kubescore

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/golang/mock/gomock"

	//"github.com/intelops/kubviz/mocks"

	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/pkg/nats/sdk"
	mocks "github.com/intelops/kubviz/pkg/nats/sdk/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/zegl/kube-score/renderer/json_v2"
)

func TestPublishKubescoreMetrics(t *testing.T) {
	// Define the report data
	report := []json_v2.ScoredObject{
		{
			ObjectName: "object1",
			// Define other fields as needed
		},
		{
			ObjectName: "object2",
			// Define other fields as needed
		},
	}

	// Initialize the MockJetStreamContextInterface
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	js := mocks.NewMockNATSClientInterface(ctrl)

	// Test case: Testing successful publishing of kube-score metrics
	t.Run("Successful publishing", func(t *testing.T) {

		// Set the mock expectation for Publish
		js.EXPECT().Publish(constants.KUBESCORE_SUBJECT, gomock.Any()).Return(nil)
		//natsCli := &sdk.NATSClient{}
		// Call the function under test
		err := publishKubescoreMetrics(report, js)

		// Assert that no error occurred during the function call
		assert.NoError(t, err)
	})
	// Test case: Error handling for Publish failure
	t.Run("Error handling for Publish failure", func(t *testing.T) {
		js.EXPECT().Publish(constants.KUBESCORE_SUBJECT, gomock.Any()).Return(errors.New("publish error"))
		//natsCli := &sdk.NATSClient{}
		err := publishKubescoreMetrics(report, js)
		assert.Error(t, err)
	})

	// Test case: Nil input report
	t.Run("Nil input report", func(t *testing.T) {
		js.EXPECT().Publish(constants.KUBESCORE_SUBJECT, gomock.Any()).Return(errors.New("publish error"))
		//natsCli := &sdk.NATSClient{}
		err := publishKubescoreMetrics(nil, js)
		assert.Error(t, err) // Assuming this is the desired behavior for nil input
	})

	// Test case: Empty report
	t.Run("Empty report", func(t *testing.T) {
		js.EXPECT().Publish(constants.KUBESCORE_SUBJECT, gomock.Any()).Return(errors.New("publish error"))
		//natsCli := &sdk.NATSClient{}
		err := publishKubescoreMetrics([]json_v2.ScoredObject{}, js)
		assert.Error(t, err) // Assuming this is the desired behavior for an empty report
	})

}
func TestExecuteCommand(t *testing.T) {
	t.Run("Successful command execution", func(t *testing.T) {
		command := "echo 'Hello, World!'"
		output, err := ExecuteCommand(command)

		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!\n", output)
	})

}

func TestPublish(t *testing.T) {
	// Mock the ExecuteCommand function
	var mockOutput = []byte(`[{"ObjectName":"test-object","TypeMeta":{"Kind":"Pod"},"ObjectMeta":{"Name":"test-pod"},"Checks":[{"ID":"check-id","Severity":"info","Message":"test message"}],"FileName":"test-file","FileRow":1}]`)
	gomonkey.ApplyFunc(ExecuteCommand, func(command string) (string, error) {
		return string(mockOutput), nil
	})
	defer gomonkey.NewPatches().Reset()

	// Create a new gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a NATSClient mock
	natsCliMock := mocks.NewMockNATSClientInterface(ctrl)

	// Patch the Publish method of NATSClient
	// patches := gomonkey.ApplyMethod(reflect.TypeOf(&sdk.NATSClient{}), "Publish", func(_ *sdk.NATSClient, subject string, data []byte) error {
	// 	return natsCliMock.Publish(subject, data)
	// })
	// defer patches.Reset()

	// Subtest for successful publish
	t.Run("Successful publish", func(t *testing.T) {
		natsCliMock.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)
		ns := "test-namespace"
		//natsCli := &sdk.NATSClient{} // Use the actual NATSClient
		err := publish(ns, natsCliMock)
		assert.NoError(t, err, "publish returned an error")
	})

	// Subtest for error in ExecuteCommand
	t.Run("Error in ExecuteCommand", func(t *testing.T) {
		// Mock ExecuteCommand to return an error
		gomonkey.ApplyFunc(ExecuteCommand, func(command string) (string, error) {
			return "", errors.New("command execution error")
		})
		defer gomonkey.NewPatches().Reset()

		ns := "test-namespace"
		natsCli := &sdk.NATSClient{} // Use the actual NATSClient
		err := publish(ns, natsCli)
		assert.Error(t, err, "publish did not return an error")
	})
}

type NATSClientWrapper struct {
	mock *mocks.MockNATSClientInterface
}

func (w *NATSClientWrapper) CreateStream(streamName string) error {
	return w.mock.CreateStream(streamName)
}

func (w *NATSClientWrapper) Publish(subject string, data []byte) error {
	return w.mock.Publish(subject, data)
}
