package trivy

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"unsafe"

	"bou.ke/monkey"
	"github.com/aquasecurity/trivy/pkg/k8s/report"
	"github.com/aquasecurity/trivy/pkg/types"
	"github.com/golang/mock/gomock"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/mocks"
	"github.com/intelops/kubviz/model"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestExecuteCommand(t *testing.T) {
	t.Run("Successful command execution", func(t *testing.T) {
		command := "echo 'Hello, World!'"
		output, err := executeCommandTrivy(command)

		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!\n", string(output))
	})

	t.Run("Command execution error", func(t *testing.T) {
		command := "non_existing_command"
		_, err := executeCommandTrivy(command)

		assert.Error(t, err)
	})
}
func TestPublishTrivyK8sReport(t *testing.T) {
	// Initialize the MockJetStreamContextInterface
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	js := mocks.NewMockJetStreamContextInterface(ctrl)

	// Define the sample consolidated report
	report := report.ConsolidatedReport{
		// populate with sample data
	}

	// Test case: Testing successful publishing of Trivy K8s report
	t.Run("Successful publishing", func(t *testing.T) {
		// Set the mock expectation for Publish
		js.EXPECT().Publish(constants.TRIVY_K8S_SUBJECT, gomock.Any()).Return(nil, nil)

		// Call the function under test
		err := PublishTrivyK8sReport(report, js)

		// Assert that no error occurred during the function call
		assert.NoError(t, err)
	})

	// Test case: Error handling for Publish failure
	t.Run("Error handling for Publish failure", func(t *testing.T) {
		// Set the mock expectation for Publish to return an error
		js.EXPECT().Publish(constants.TRIVY_K8S_SUBJECT, gomock.Any()).Return(nil, errors.New("publish error"))

		// Call the function under test
		err := PublishTrivyK8sReport(report, js)

		// Assert that an error occurred during the function call
		assert.Error(t, err)
	})

}
func TestRunTrivyK8sClusterScan(t *testing.T) {
	// Mock the executeCommandTrivy function
	var mockOutput = []byte(`{"ObjectName":"test-object","TypeMeta":{"Kind":"Pod"},"ObjectMeta":{"Name":"test-pod"},"Checks":[{"ID":"check-id","Severity":"info","Message":"test message"}],"FileName":"test-file","FileRow":1}`)
	monkey.Patch(executeCommandTrivy, func(command string) ([]byte, error) {
		return mockOutput, nil
	})
	defer monkey.Unpatch(executeCommandTrivy)

	// Mock the os.MkdirAll function
	monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
		// Do nothing and return nil (assuming that the directory creation is successful)
		return nil
	})
	defer monkey.Unpatch(os.MkdirAll)

	// Create a new gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a JetStreamContext mock
	jsMock := mocks.NewMockJetStreamContextInterface(ctrl)

	// Test case: Successful Trivy scan
	t.Run("Successful scan", func(t *testing.T) {
		// Set the mock expectation for PublishTrivyK8sReport
		jsMock.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil, nil)

		// Call the function under test
		err := RunTrivyK8sClusterScan(jsMock)
		assert.NoError(t, err)
	})

	// Test case: Error in executeCommandTrivy
	// Test case: Error in executeCommandTrivy
	t.Run("Error in executeCommandTrivy", func(t *testing.T) {
		// Mock executeCommandTrivy to return an error
		monkey.Patch(executeCommandTrivy, func(command string) ([]byte, error) {
			return nil, errors.New("command execution error")
		})
		defer monkey.Unpatch(executeCommandTrivy)

		// Publish should not be called since executeCommandTrivy failed
		jsMock.EXPECT().Publish(gomock.Any(), gomock.Any()).Times(0)

		// Call the function under test
		err := RunTrivyK8sClusterScan(jsMock)
		assert.Error(t, err)
	})
	// Test case: Empty output from executeCommandTrivy
	// Test case: Empty output from executeCommandTrivy
	t.Run("Empty output from executeCommandTrivy", func(t *testing.T) {
		// Mock executeCommandTrivy to return empty output
		monkey.Patch(executeCommandTrivy, func(command string) ([]byte, error) {
			return []byte{}, nil
		})
		defer monkey.Unpatch(executeCommandTrivy)

		// Publish should not be called since output is empty
		jsMock.EXPECT().Publish(gomock.Any(), gomock.Any()).Times(0)

		// Call the function under test
		err := RunTrivyK8sClusterScan(jsMock)
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	})
	// Test case: Invalid JSON output from executeCommandTrivy
	// Test case: Invalid JSON output from executeCommandTrivy
	t.Run("Invalid JSON output from executeCommandTrivy", func(t *testing.T) {
		// Mock executeCommandTrivy to return invalid JSON output
		monkey.Patch(executeCommandTrivy, func(command string) ([]byte, error) {
			return []byte("invalid json"), nil
		})
		defer monkey.Unpatch(executeCommandTrivy)

		// Publish should not be called since JSON is invalid
		jsMock.EXPECT().Publish(gomock.Any(), gomock.Any()).Times(0)

		// Call the function under test
		err := RunTrivyK8sClusterScan(jsMock)
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	})
	// Test case: Error in Publish
	t.Run("Error in Publish", func(t *testing.T) {
		// Mock executeCommandTrivy to return valid output
		monkey.Patch(executeCommandTrivy, func(command string) ([]byte, error) {
			return []byte(`{"ObjectName":"test-object","TypeMeta":{"Kind":"Pod"},"ObjectMeta":{"Name":"test-pod"},"Checks":[{"ID":"check-id","Severity":"info","Message":"test message"}],"FileName":"test-file","FileRow":1}`), nil
		})
		defer monkey.Unpatch(executeCommandTrivy)

		// Mock Publish to return an error
		jsMock.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil, errors.New("publish error"))

		// Call the function under test
		err := RunTrivyK8sClusterScan(jsMock)
		assert.Error(t, err)
	})
}
func TestPublishTrivySbomReport(t *testing.T) {
	// Create a new gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a JetStreamContext mock
	jsMock := mocks.NewMockJetStreamContextInterface(ctrl)

	// Define a sample report
	report := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	// Test case: Successful publishing of Trivy SBOM report
	t.Run("Successful publishing", func(t *testing.T) {
		// Set the mock expectation for Publish
		jsMock.EXPECT().Publish(constants.TRIVY_SBOM_SUBJECT, gomock.Any()).Return(nil, nil)

		// Call the function under test
		err := PublishTrivySbomReport(report, jsMock)

		// Assert that no error occurred during the function call
		assert.NoError(t, err)
	})

	// Test case: Error handling for Publish failure
	t.Run("Error handling for Publish failure", func(t *testing.T) {
		// Set the mock expectation for Publish to return an error
		jsMock.EXPECT().Publish(constants.TRIVY_SBOM_SUBJECT, gomock.Any()).Return(nil, errors.New("publish error"))

		// Call the function under test
		err := PublishTrivySbomReport(report, jsMock)

		// Assert that an error occurred during the function call
		assert.Error(t, err)
	})

	// Test case: Error marshalling the report
	t.Run("Error marshalling report", func(t *testing.T) {
		// Mocking json.Marshal to return an error
		monkey.Patch(json.Marshal, func(v interface{}) ([]byte, error) {
			return nil, errors.New("marshal error")
		})
		defer monkey.Unpatch(json.Marshal)

		// Call the function under test
		err := PublishTrivySbomReport(report, jsMock)

		// Assert that an error occurred during the function call
		assert.Error(t, err)
	})
}
func TestExecuteTrivyImageScan(t *testing.T) {
	// Create a new gomock controller

	t.Run("Successful command execution", func(t *testing.T) {
		command := "echo 'Hello, World!'"
		output, err := executeTrivyImage(command)

		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!\n", string(output))
	})

	t.Run("Command execution error", func(t *testing.T) {
		command := "non_existing_command"
		_, err := executeTrivyImage(command)

		assert.Error(t, err)
	})
}

func TestPublishImageScanReports(t *testing.T) {
	// Create a new gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a JetStreamContext mock
	jsMock := mocks.NewMockJetStreamContextInterface(ctrl)

	// Define a sample report
	report := types.Report{
		// Define your report structure here
	}

	// Test case: Successful publishing of Trivy image scan report
	t.Run("Successful publishing", func(t *testing.T) {
		// Set the mock expectation for Publish
		jsMock.EXPECT().Publish(constants.TRIVY_IMAGE_SUBJECT, gomock.Any()).Return(nil, nil)

		// Call the function under test
		err := PublishImageScanReports(report, jsMock)

		// Assert that no error occurred during the function call
		assert.NoError(t, err)
	})

	// Test case: Error handling for Publish failure
	t.Run("Error handling for Publish failure", func(t *testing.T) {
		// Set the mock expectation for Publish to return an error
		jsMock.EXPECT().Publish(constants.TRIVY_IMAGE_SUBJECT, gomock.Any()).Return(nil, errors.New("publish error"))

		// Call the function under test
		err := PublishImageScanReports(report, jsMock)

		// Assert that an error occurred during the function call
		assert.Error(t, err)
	})

	// Test case: Error marshalling the report

}
func TestExecuteCommandSbom(t *testing.T) {
	// Create a new gomock controller

	t.Run("Successful command execution", func(t *testing.T) {
		command := "echo 'Hello, World!'"
		output, err := executeCommandSbom(command)

		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!\n", string(output))
	})

	t.Run("Command execution error", func(t *testing.T) {
		command := "non_existing_command"
		_, err := executeCommandSbom(command)

		assert.Error(t, err)
	})
}
func TestRunTrivySbomScan(t *testing.T) {
	// Replace the ListImagesforSbom function with a mock implementation
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a JetStreamContext mock
	jsMock := mocks.NewMockJetStreamContextInterface(ctrl)
	monkey.Patch(ListImagesforSbom, func(config *rest.Config) ([]model.RunningImage, error) {
		return []model.RunningImage{{PullableImage: "image1"}}, nil
	})
	// Replace the ExecuteCommandSbom function with a mock implementation
	monkey.Patch(executeCommandSbom, func(command string) ([]byte, error) {
		return []byte(`{...}`), nil
	})
	monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
		// Do nothing and return nil (assuming that the directory creation is successful)
		return nil
	})
	defer monkey.Unpatch(os.MkdirAll)
	// Run your test
	err := RunTrivySbomScan(&rest.Config{}, jsMock)

	assert.NoError(t, err)

	// Restore the original functions
	monkey.UnpatchAll()
}

func TestRunTrivyImageScans(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jsMock := mocks.NewMockJetStreamContextInterface(ctrl)

	images := []model.RunningImage{
		{PullableImage: "image1"},
		{PullableImage: "image2"},
	}

	// Patch the ListImages function to return your predefined images
	monkey.Patch(ListImages, func(config *rest.Config) ([]model.RunningImage, error) {
		return images, nil
	})

	// Patch the ExecuteTrivyImage function to return a predefined result
	monkey.Patch(executeTrivyImage, func(command string) ([]byte, error) {
		return []byte(`{...}`), nil
	})

	// Patch os.MkdirAll to return an error to simulate a failure
	monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
		// Do nothing and return nil (assuming that the directory creation is successful)
		return nil
	})
	defer monkey.Unpatch(os.MkdirAll)
	// Test case for successful execution
	err := RunTrivyImageScans(&rest.Config{}, jsMock)
	assert.NoError(t, err)

	// Test case for failure in ListImages
	monkey.Patch(ListImages, func(config *rest.Config) ([]model.RunningImage, error) {
		return nil, errors.New("error listing images")
	})
	err = RunTrivyImageScans(&rest.Config{}, jsMock)
	assert.Error(t, err)

	// Test case for failure in ExecuteTrivyImage
	monkey.Patch(executeTrivyImage, func(command string) ([]byte, error) {
		return nil, errors.New("error executing trivy")
	})
	monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
		// Do nothing and return nil (assuming that the directory creation is successful)
		return nil
	})
	defer monkey.Unpatch(os.MkdirAll)
	err = RunTrivyImageScans(&rest.Config{}, jsMock)
	assert.Error(t, err)

	// Unpatch the patched functions
	monkey.UnpatchAll()
}
func NewFakeClientset() *kubernetes.Clientset {
	fakeClientset := fake.NewSimpleClientset()
	return (*kubernetes.Clientset)(unsafe.Pointer(fakeClientset))
}
