package handler

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mock_main "github.com/intelops/kubviz/agent/container/pkg/clients"
	"github.com/stretchr/testify/assert"
)

func TestGetLiveness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := &APIHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	app.GetStatus(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPostEventAzureContainer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockNatsClient := mock_main.NewMockNATSClientInterface(mockCtrl)
	app := &APIHandler{
		conn: mockNatsClient,
	}

	tests := []struct {
		name           string
		headerEvent    string
		bodyData       []byte
		expectedLog    string
		expectedStatus int
		mockPublishErr error
	}{
		{
			name:           "Valid request",
			headerEvent:    "event",
			bodyData:       []byte(`{"id":"123","timestamp":"2024-06-10T10:00:00Z","action":"push","target":{"mediaType":"application/vnd.docker.distribution.manifest.v2+json","size":123,"digest":"sha256:1234567890abcdef","length":123,"repository":"repo","tag":"latest"},"request":{"id":"456","host":"localhost","method":"GET","useragent":"curl"}}`),
			expectedLog:    "Received event from Azure Container Registry",
			expectedStatus: http.StatusOK,
			mockPublishErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the recorder for each test case
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set the request body and header
			req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.bodyData))
			req.Header.Set("X-Event", tt.headerEvent)
			c.Request = req

			// Set the expectation on the mock
			if tt.mockPublishErr != nil {
				mockNatsClient.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(tt.mockPublishErr)
			} else {
				mockNatsClient.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)
			}
			var logOutput bytes.Buffer
			log.SetOutput(&logOutput)
			defer log.SetOutput(os.Stderr)
			// Perform the request
			app.PostEventAzureContainer(c)
			// Verify the log output using strings.Contains
			logStr := logOutput.String()
			assert.Contains(t, logStr, tt.expectedLog, "log output should contain the expected log")

			// Check the response status code
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestPostEventDockerHub(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockNatsClient := mock_main.NewMockNATSClientInterface(mockCtrl)
	app := &APIHandler{
		conn: mockNatsClient,
	}

	// Define test cases
	tests := []struct {
		name           string
		headerEvent    string
		bodyData       []byte
		expectedLog    string
		expectedStatus int
		mockPublishErr error
	}{
		{
			name:           "Valid request",
			headerEvent:    "event",
			bodyData:       []byte(`{"key": "value"}`),
			expectedLog:    "Received event from docker artifactory:",
			expectedStatus: http.StatusOK,
			mockPublishErr: nil,
		},
		{
			name:           "Empty body",
			headerEvent:    "event",
			bodyData:       []byte{},
			expectedLog:    "error reading the request body",
			expectedStatus: http.StatusOK,
			mockPublishErr: nil,
		},
		{
			name:           "Error publishing to NATS",
			headerEvent:    "event",
			bodyData:       []byte(`{"key": "value"}`),
			expectedLog:    "error while publishing to nats",
			expectedStatus: http.StatusInternalServerError,
			mockPublishErr: errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the recorder for each test case
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set the request body and header
			req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.bodyData))
			req.Header.Set("X-Event", tt.headerEvent)
			c.Request = req

			// Set the mock expectation only if the body is not empty
			if len(tt.bodyData) > 0 {
				mockNatsClient.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(tt.mockPublishErr)
			}

			var logOutput bytes.Buffer
			log.SetOutput(&logOutput)
			defer log.SetOutput(os.Stderr)

			// Perform the request
			app.PostEventDockerHub(c)

			// Verify the log output using strings.Contains
			logStr := logOutput.String()
			assert.Contains(t, logStr, tt.expectedLog, "log output should contain the expected log")

			// Check the response status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Log the error message and request body for debugging
			if w.Code != tt.expectedStatus {
				t.Log("Response body:", w.Body.String())
				t.Log("Request body:", string(tt.bodyData))
			}

		})
	}
}

func TestPostEventJfrogContainer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockNatsClient := mock_main.NewMockNATSClientInterface(mockCtrl)
	app := &APIHandler{
		conn: mockNatsClient,
	}

	tests := []struct {
		name           string
		headerEvent    string
		bodyData       []byte
		expectedLog    string
		expectedStatus int
		mockPublishErr error
	}{
		{
			name:           "Valid request",
			headerEvent:    "event",
			bodyData:       []byte(`{"domain":"domain","event_type":"event","data":{"repo_key":"key","path":"path","name":"name","sha256":"sha","size":123,"image_name":"image","tag":"tag"},"subscription_key":"sub","jpd_origin":"origin","source":"source"}`),
			expectedLog:    "Received event from jfrog Container Registry",
			expectedStatus: http.StatusOK,
			mockPublishErr: nil,
		},
		{
			name:           "Empty body",
			headerEvent:    "event",
			bodyData:       []byte{},
			expectedLog:    "error reading the request body",
			expectedStatus: http.StatusOK,
			mockPublishErr: nil,
		},
		{
			name:           "Error publishing to NATS",
			headerEvent:    "event",
			bodyData:       []byte(`{"domain":"domain","event_type":"event","data":{"repo_key":"key","path":"path","name":"name","sha256":"sha","size":123,"image_name":"image","tag":"tag"},"subscription_key":"sub","jpd_origin":"origin","source":"source"}`),
			expectedLog:    "Received event from jfrog Container",
			expectedStatus: http.StatusInternalServerError,
			mockPublishErr: errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.bodyData))
			req.Header.Set("X-Event", tt.headerEvent)
			c.Request = req

			// Set the mock expectation only if the body is not empty
			if len(tt.bodyData) > 0 {
				mockNatsClient.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(tt.mockPublishErr)
			}
			var logOutput bytes.Buffer
			log.SetOutput(&logOutput)
			defer log.SetOutput(os.Stderr)

			app.PostEventJfrogContainer(c)

			logStr := logOutput.String()
			assert.Contains(t, logStr, tt.expectedLog, "log output should contain the expected log")

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code != tt.expectedStatus {
				t.Log("Response body:", w.Body.String())
				t.Log("Request body:", string(tt.bodyData))
			}
		})
	}
}

func TestPostEventQuayContainer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockNatsClient := mock_main.NewMockNATSClientInterface(mockCtrl)
	app := &APIHandler{
		conn: mockNatsClient,
	}

	tests := []struct {
		name           string
		headerEvent    string
		bodyData       []byte
		expectedLog    string
		expectedStatus int
		mockPublishErr error
	}{
		{
			name:           "Valid request",
			headerEvent:    "event",
			bodyData:       []byte(`{"name":"name","repository":"repo","namespace":"namespace","docker_url":"url","homepage":"home","updated_tags":["tag1","tag2"]}`),
			expectedLog:    "Received event from Quay Container Registry",
			expectedStatus: http.StatusOK,
			mockPublishErr: nil,
		},
		{
			name:           "Empty body",
			headerEvent:    "event",
			bodyData:       []byte{},
			expectedLog:    "error reading the request body",
			expectedStatus: http.StatusOK,
			mockPublishErr: nil,
		},
		{
			name:           "Error publishing to NATS",
			headerEvent:    "event",
			bodyData:       []byte(`{"name":"name","repository":"repo","namespace":"namespace","docker_url":"url","homepage":"home","updated_tags":["tag1","tag2"]}`),
			expectedLog:    "Received event from Quay Container",
			expectedStatus: http.StatusInternalServerError,
			mockPublishErr: errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.bodyData))
			req.Header.Set("X-Event", tt.headerEvent)
			c.Request = req

			// Set the mock expectation only if the body is not empty
			if len(tt.bodyData) > 0 {
				mockNatsClient.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(tt.mockPublishErr)
			}

			var logOutput bytes.Buffer
			log.SetOutput(&logOutput)
			defer log.SetOutput(os.Stderr)

			app.PostEventQuayContainer(c)

			logStr := logOutput.String()
			assert.Contains(t, logStr, tt.expectedLog, "log output should contain the expected log")

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code != tt.expectedStatus {
				t.Log("Response body:", w.Body.String())
				t.Log("Request body:", string(tt.bodyData))
			}
		})
	}
}
