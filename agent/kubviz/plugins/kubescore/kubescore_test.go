package kubescore

import (
	"errors"
	"testing"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/mocks"
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

	js := mocks.NewMockJetStreamContextInterface(ctrl)

	// Test case: Testing successful publishing of kube-score metrics
	t.Run("Successful publishing", func(t *testing.T) {

		// Set the mock expectation for Publish
		js.EXPECT().Publish(constants.KUBESCORE_SUBJECT, gomock.Any()).Return(nil, nil)

		// Call the function under test
		err := publishKubescoreMetrics(report, js)

		// Assert that no error occurred during the function call
		assert.NoError(t, err)
	})
	// Test case: Error handling for Publish failure
	t.Run("Error handling for Publish failure", func(t *testing.T) {
		js.EXPECT().Publish(constants.KUBESCORE_SUBJECT, gomock.Any()).Return(nil, errors.New("publish error"))
		err := publishKubescoreMetrics(report, js)
		assert.Error(t, err)
	})

	// Test case: Nil input report
	t.Run("Nil input report", func(t *testing.T) {
		js.EXPECT().Publish(constants.KUBESCORE_SUBJECT, gomock.Any()).Return(nil, errors.New("publish error"))

		err := publishKubescoreMetrics(nil, js)
		assert.Error(t, err) // Assuming this is the desired behavior for nil input
	})

	// Test case: Empty report
	t.Run("Empty report", func(t *testing.T) {
		js.EXPECT().Publish(constants.KUBESCORE_SUBJECT, gomock.Any()).Return(nil, errors.New("publish error"))

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

	// t.Run("Command execution error", func(t *testing.T) {
	// 	command := "non_existing_command"
	// 	_, err := ExecuteCommand(command)

	// 	assert.Error(t, err)
	// })

}

func TestPublish(t *testing.T) {
	// Mock the ExecuteCommand function
	var mockOutput = []byte(`[{"ObjectName":"test-object","TypeMeta":{"Kind":"Pod"},"ObjectMeta":{"Name":"test-pod"},"Checks":[{"ID":"check-id","Severity":"info","Message":"test message"}],"FileName":"test-file","FileRow":1}]`)
	monkey.Patch(ExecuteCommand, func(command string) (string, error) {
		return string(mockOutput), nil
	})
	defer monkey.Unpatch(ExecuteCommand)

	// Create a new gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a JetStreamContext mock
	jsMock := mocks.NewMockJetStreamContextInterface(ctrl)

	// Subtest for successful publish
	t.Run("Successful publish", func(t *testing.T) {
		jsMock.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil, nil)
		ns := "test-namespace"
		err := publish(ns, jsMock)
		if err != nil {
			t.Errorf("publish returned an error: %v", err)
		}
	})

	// Subtest for error in ExecuteCommand
	t.Run("Error in ExecuteCommand", func(t *testing.T) {
		// Mock ExecuteCommand to return an error
		monkey.Patch(ExecuteCommand, func(command string) (string, error) {
			return "", errors.New("command execution error")
		})
		defer monkey.Unpatch(ExecuteCommand)

		ns := "test-namespace"
		err := publish(ns, jsMock)
		if err == nil {
			t.Errorf("publish did not return an error")
		}

		// Since ExecuteCommand failed, Publish should not be called
		jsMock.EXPECT().Publish(gomock.Any(), gomock.Any()).Times(0)
	})
}
