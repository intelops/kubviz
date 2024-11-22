package application

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/intelops/kubviz/agent/git/api"
	"github.com/intelops/kubviz/agent/git/pkg/clients"
	"github.com/intelops/kubviz/agent/git/pkg/clients/mocks"
	"github.com/intelops/kubviz/agent/git/pkg/config"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetApiDocs(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create an instance of the Application struct
	app := &Application{}

	// Define the test cases
	tests := []struct {
		name         string
		mockResponse *openapi3.T
		mockError    error
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success",
			mockResponse: &openapi3.T{
				OpenAPI: "3.0.0",
				Info: &openapi3.Info{
					Title:   "Sample API",
					Version: "1.0.0",
				},
				Paths: openapi3.Paths{},
			},
			mockError:    nil,
			expectedCode: http.StatusOK,
			expectedBody: `{"openapi":"3.0.0","info":{"title":"Sample API","version":"1.0.0"},"paths":{}}`,
		},
		{
			name:         "Error",
			mockResponse: nil,
			mockError:    errors.New("error fetching swagger"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Patch the GetSwagger function
			patch := gomonkey.ApplyFunc(api.GetSwagger, func() (*openapi3.T, error) {
				return tt.mockResponse, tt.mockError
			})
			defer patch.Reset()

			// Create a new Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Call the GetApiDocs method
			app.GetApiDocs(c)

			// Verify the response
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedCode == http.StatusOK {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}
		})
	}
}

func TestGetLiveness(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create an instance of the Application struct
	app := &Application{}

	// Create a new Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Call the GetLiveness method
	app.GetLiveness(c)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
}
func TestNew(t *testing.T) {
	// Test case: valid configuration and NATS connection
	conf := &config.Config{Port: 8080}
	conn := &clients.NATSContext{}
	app := New(conf, conn)
	if app.Config != conf {
		t.Errorf("Expected Config to be %v, got %v", conf, app.Config)
	}
	if app.conn != conn {
		t.Errorf("Expected conn to be %v, got %v", conn, app.conn)
	}
	if app.server.Addr != ":8081" {
		t.Errorf("Expected server.Addr to be :8081, got %s", app.server.Addr)
	}
	if app.server.Handler == nil {
		t.Error("Expected server.Handler to be non-nil")
	}
	if app.server.IdleTimeout != time.Minute {
		t.Errorf("Expected server.IdleTimeout to be %v, got %v", time.Minute, app.server.IdleTimeout)
	}
	if app.server.ReadTimeout != 10*time.Second {
		t.Errorf("Expected server.ReadTimeout to be %v, got %v", 10*time.Second, app.server.ReadTimeout)
	}
	if app.server.WriteTimeout != 30*time.Second {
		t.Errorf("Expected server.WriteTimeout to be %v, got %v", 30*time.Second, app.server.WriteTimeout)
	}

	// Test case: nil configuration
	app = New(nil, conn)
	if app.Config != nil {
		t.Errorf("Expected Config to be nil, got %v", app.Config)
	}

	// Test case: nil NATS connection
	app = New(conf, nil)
	if app.conn != nil {
		t.Errorf("Expected conn to be nil, got %v", app.conn)
	}
}

func TestStart(t *testing.T) {
	// Create an instance of the Application struct
	app := &Application{
		server: &http.Server{Addr: ":8080"}, // Initialize app.server with a valid http.Server instance
	}

	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with a success status code
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Start the server in a goroutine
	go func() {
		// Pass the test server's URL to ListenAndServe
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Fatalf("Server closed, reason: %v", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	// Make a request to the server
	resp, err := http.Get(testServer.URL)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	// Verify the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

type MockNATSContext struct {
	clients.NATSContext
	*http.Server
	mock.Mock
}

func (m *MockNATSContext) Close() {
	m.Called()
}

func (m *MockNATSContext) CreateStream() (nats.JetStreamContext, error) {
	args := m.Called()
	return args.Get(0).(nats.JetStreamContext), args.Error(1)
}

func (m *MockNATSContext) Publish(metric []byte, repo string, eventkey model.EventKey, eventvalue model.EventValue) error {
	args := m.Called(metric, repo, eventkey, eventvalue)
	return args.Error(0)
}

// Helper type to capture log output

// Mock the connection interface
func TestPostGitea(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock NATSContext
	mockConn := mocks.NewMockNATSClientInterface(ctrl)
	// Create an instance of the Application struct
	app := &Application{conn: mockConn}

	// Define the test cases
	tests := []struct {
		name           string
		headerEvent    string
		bodyData       []byte
		expectedLog    string
		expectedStatus int
		mockPublishErr error
	}{
		{
			name:           "Success",
			headerEvent:    "push",
			bodyData:       []byte(`{"ref": "refs/heads/main"}`),
			expectedLog:    `GITEA DATA: "{\"ref\": \"refs/heads/main\"}"`,
			expectedStatus: http.StatusOK,
			mockPublishErr: nil,
		},
		{
			name:           "Missing Event Header",
			headerEvent:    "",
			bodyData:       nil,
			expectedLog:    "error getting the gitea event from header",
			expectedStatus: http.StatusBadRequest,
			mockPublishErr: nil,
		},
		{
			name:           "Publish Error",
			headerEvent:    "push",
			bodyData:       []byte(`{"ref": "refs/heads/main"}`),
			expectedLog:    `GITEA DATA: "{\"ref\": \"refs/heads/main\"}"`,
			expectedStatus: http.StatusInternalServerError,
			mockPublishErr: errors.New("publish error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock response
			if tt.headerEvent != "" {
				mockConn.EXPECT().Publish(tt.bodyData, string(model.GiteaProvider), model.GiteaHeader, model.EventValue(tt.headerEvent)).Return(tt.mockPublishErr).Times(1)
			}

			// Create a new Gin context with the necessary headers and body
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.bodyData))
			c.Request.Header.Set(string(model.GiteaHeader), tt.headerEvent)

			// Capture logs
			var logOutput bytes.Buffer
			log.SetOutput(&logOutput)
			defer log.SetOutput(os.Stderr)

			// Call the PostGitea method
			app.PostGitea(c)

			// Verify the log output using strings.Contains
			logStr := logOutput.String()
			assert.Contains(t, logStr, tt.expectedLog, "log output should contain the expected log")

			// Verify the response status code
			if tt.mockPublishErr != nil {
				assert.Equal(t, tt.expectedStatus, http.StatusInternalServerError)
			} else if tt.headerEvent == "" {
				assert.Equal(t, tt.expectedStatus, http.StatusBadRequest)
			} else {
				assert.Equal(t, tt.expectedStatus, w.Code)
			}
		})
	}
}
func TestPostAzure(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock NATSContext
	mockConn := mocks.NewMockNATSClientInterface(ctrl)
	// Create an instance of the Application struct
	app := &Application{conn: mockConn}

	// Define the test cases
	tests := []struct {
		name           string
		bodyData       []byte
		eventType      string
		expectedLog    string
		expectedStatus int
		mockPublishErr error
	}{
		{
			name:           "Success",
			bodyData:       []byte(`{"eventType": "push"}`),
			eventType:      "push",
			expectedLog:    `AZURE DATA: "{\"eventType\": \"push\"}"`,
			expectedStatus: http.StatusOK,
			mockPublishErr: nil,
		},
		{
			name:           "Missing EventType",
			bodyData:       []byte(`{}`),
			eventType:      "",
			expectedLog:    "Error Reading Request Body",
			expectedStatus: http.StatusInternalServerError,
			mockPublishErr: nil,
		},
		{
			name:           "Unmarshal Error",
			bodyData:       []byte(`invalid json`),
			eventType:      "",
			expectedLog:    "Error Reading Request Body",
			expectedStatus: http.StatusInternalServerError,
			mockPublishErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock response
			if tt.eventType != "" {
				mockConn.EXPECT().Publish(tt.bodyData, string(model.AzureDevopsProvider), model.AzureHeader, model.EventValue(tt.eventType)).Return(tt.mockPublishErr).Times(1)
			}

			// Create a new Gin context with the necessary headers and body
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.bodyData))

			// Capture logs
			var logOutput bytes.Buffer
			log.SetOutput(&logOutput)
			defer log.SetOutput(os.Stderr)

			// Call the PostAzure method
			app.PostAzure(c)

			// Verify the log output using strings.Contains
			logStr := logOutput.String()
			assert.Contains(t, logStr, tt.expectedLog, "log output should contain the expected log")

			// Verify the response status code
			if tt.eventType == "" {
				assert.Equal(t, tt.expectedStatus, http.StatusInternalServerError)
			} else {
				assert.Equal(t, tt.expectedStatus, w.Code)
			}
		})
	}
}
func TestPostGithub(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock NATSContext
	mockConn := mocks.NewMockNATSClientInterface(ctrl)
	// Create an instance of the Application struct
	app := &Application{conn: mockConn}

	// Define the test cases
	tests := []struct {
		name           string
		headerEvent    string
		bodyData       []byte
		expectedLog    string
		expectedStatus int
		mockPublishErr error
	}{
		{
			name:           "Success",
			headerEvent:    "push",
			bodyData:       []byte(`{"ref": "refs/heads/main"}`),
			expectedLog:    `GITHUB DATA: "{\"ref\": \"refs/heads/main\"}"`,
			expectedStatus: http.StatusOK,
			mockPublishErr: nil,
		},
		{
			name:           "Missing Event Header",
			headerEvent:    "",
			bodyData:       nil,
			expectedLog:    "error getting the github event from header",
			expectedStatus: http.StatusBadRequest,
			mockPublishErr: nil,
		},
		{
			name:           "Publish Error",
			headerEvent:    "push",
			bodyData:       []byte(`{"ref": "refs/heads/main"}`),
			expectedLog:    `GITHUB DATA: "{\"ref\": \"refs/heads/main\"}"`,
			expectedStatus: http.StatusInternalServerError,
			mockPublishErr: errors.New("publish error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock response
			if tt.headerEvent != "" {
				mockConn.EXPECT().Publish(tt.bodyData, string(model.GithubProvider), model.GithubHeader, model.EventValue(tt.headerEvent)).Return(tt.mockPublishErr).Times(1)
			}

			// Create a new Gin context with the necessary headers and body
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.bodyData))
			c.Request.Header.Set(string(model.GithubHeader), tt.headerEvent)

			// Capture logs
			var logOutput bytes.Buffer
			log.SetOutput(&logOutput)
			defer log.SetOutput(os.Stderr)

			// Call the PostGithub method
			app.PostGithub(c)

			// Verify the log output using strings.Contains
			logStr := logOutput.String()
			assert.Contains(t, logStr, tt.expectedLog, "log output should contain the expected log")

			// Verify the response status code
			if tt.mockPublishErr != nil {
				assert.Equal(t, tt.expectedStatus, http.StatusInternalServerError)
			} else if tt.headerEvent == "" {
				assert.Equal(t, tt.expectedStatus, http.StatusBadRequest)
			} else {
				assert.Equal(t, tt.expectedStatus, w.Code)
			}
		})
	}
}
func TestPostGitlab(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock NATSContext
	mockConn := mocks.NewMockNATSClientInterface(ctrl)
	// Create an instance of the Application struct
	app := &Application{conn: mockConn}

	// Define the test cases
	tests := []struct {
		name           string
		headerEvent    string
		bodyData       []byte
		expectedLog    string
		expectedStatus int
		mockPublishErr error
	}{
		{
			name:           "Success",
			headerEvent:    "push",
			bodyData:       []byte(`{"ref": "refs/heads/main"}`),
			expectedLog:    `GITLAB DATA: "{\"ref\": \"refs/heads/main\"}"`,
			expectedStatus: http.StatusOK,
			mockPublishErr: nil,
		},
		{
			name:           "Missing Event Header",
			headerEvent:    "",
			bodyData:       nil,
			expectedLog:    "error getting the gitlab event from header",
			expectedStatus: http.StatusBadRequest,
			mockPublishErr: nil,
		},
		{
			name:           "Publish Error",
			headerEvent:    "push",
			bodyData:       []byte(`{"ref": "refs/heads/main"}`),
			expectedLog:    `GITLAB DATA: "{\"ref\": \"refs/heads/main\"}"`,
			expectedStatus: http.StatusInternalServerError,
			mockPublishErr: errors.New("publish error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock response
			if tt.headerEvent != "" {
				mockConn.EXPECT().Publish(tt.bodyData, string(model.GitlabProvider), model.GitlabHeader, model.EventValue(tt.headerEvent)).Return(tt.mockPublishErr).Times(1)
			}

			// Create a new Gin context with the necessary headers and body
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.bodyData))
			c.Request.Header.Set(string(model.GitlabHeader), tt.headerEvent)

			// Capture logs
			var logOutput bytes.Buffer
			log.SetOutput(&logOutput)
			defer log.SetOutput(os.Stderr)

			// Call the PostGitlab method
			app.PostGitlab(c)

			// Verify the log output using strings.Contains
			logStr := logOutput.String()
			assert.Contains(t, logStr, tt.expectedLog, "log output should contain the expected log")

			// Verify the response status code
			if tt.mockPublishErr != nil {
				assert.Equal(t, tt.expectedStatus, http.StatusInternalServerError)
			} else if tt.headerEvent == "" {
				assert.Equal(t, tt.expectedStatus, http.StatusBadRequest)
			} else {
				assert.Equal(t, tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestPostBitbucket(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock NATSContext
	mockConn := mocks.NewMockNATSClientInterface(ctrl)
	// Create an instance of the Application struct
	app := &Application{conn: mockConn}

	// Define the test cases
	tests := []struct {
		name           string
		headerEvent    string
		bodyData       []byte
		expectedLog    string
		expectedStatus int
		mockPublishErr error
	}{
		{
			name:           "Success",
			headerEvent:    "push",
			bodyData:       []byte(`{"ref": "refs/heads/main"}`),
			expectedLog:    `BITBUCKET DATA: "{\"ref\": \"refs/heads/main\"}"`,
			expectedStatus: http.StatusOK,
			mockPublishErr: nil,
		},
		{
			name:           "Missing Event Header",
			headerEvent:    "",
			bodyData:       nil,
			expectedLog:    "error getting the bitbucket event from header",
			expectedStatus: http.StatusBadRequest,
			mockPublishErr: nil,
		},
		{
			name:           "Publish Error",
			headerEvent:    "push",
			bodyData:       []byte(`{"ref": "refs/heads/main"}`),
			expectedLog:    `BITBUCKET DATA: "{\"ref\": \"refs/heads/main\"}"`,
			expectedStatus: http.StatusInternalServerError,
			mockPublishErr: errors.New("publish error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock response
			if tt.headerEvent != "" {
				mockConn.EXPECT().Publish(tt.bodyData, string(model.BitBucketProvider), model.BitBucketHeader, model.EventValue(tt.headerEvent)).Return(tt.mockPublishErr).Times(1)
			}

			// Create a new Gin context with the necessary headers and body
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.bodyData))
			c.Request.Header.Set(string(model.BitBucketHeader), tt.headerEvent)

			// Capture logs
			var logOutput bytes.Buffer
			log.SetOutput(&logOutput)
			defer log.SetOutput(os.Stderr)

			// Call the PostBitbucket method
			app.PostBitbucket(c)

			// Verify the log output using strings.Contains
			logStr := logOutput.String()
			assert.Contains(t, logStr, tt.expectedLog, "log output should contain the expected log")

			// Verify the response status code
			if tt.mockPublishErr != nil {
				assert.Equal(t, tt.expectedStatus, http.StatusInternalServerError)
			} else if tt.headerEvent == "" {
				assert.Equal(t, tt.expectedStatus, http.StatusBadRequest)
			} else {
				assert.Equal(t, tt.expectedStatus, w.Code)
			}
		})
	}
}
func TestClose(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock NATSContext
	mockConn := mocks.NewMockNATSClientInterface(ctrl) // Expect the Close method to be called
	mockConn.EXPECT().Close().Times(1)

	// Create a mock http.Server
	mockServer := &http.Server{}
	// Expect the Shutdown method to be called
	//mockServer.EXPECT().Shutdown(gomock.Any()).Return(nil).Times(1)

	// Create an instance of the Application struct
	app := &Application{
		conn:   mockConn,
		server: mockServer,
	}

	// Call the Close method
	app.Close()

	// Verify that the expectations were met
	// This is optional depending on your needs
	// You can use assert from the testify package or similar
}
