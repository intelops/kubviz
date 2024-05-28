package kuberhealthy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey"
	"github.com/intelops/kubviz/agent/config"
	"github.com/intelops/kubviz/pkg/nats/sdk"
	khstatev1 "github.com/kuberhealthy/kuberhealthy/v2/pkg/apis/khstate/v1"
	"github.com/kuberhealthy/kuberhealthy/v2/pkg/health"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//Before running this test, go to agent/config/config.go file
//make sure set KHConfig put a default url and 0.01s poll interval
//example below
//type KHConfig struct {
//KuberhealthyURL string        `envconfig:"KUBERHEALTHY_URL" required:"true" default:"test.com"`
//PollInterval    time.Duration `envconfig:"POLL_INTERVAL" default:"0.01s"`
//}

func TestStartKuberhealthy(t *testing.T) {
	cases := []struct {
		name       string
		wantErr    bool
		publishErr bool
	}{
		{"success", false, false},
		{"failure in pollAndPublishKuberhealthy", false, true},
	}
	for _, tt := range cases {
		log.Println("Running test case: ", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			//mockJS := &MockJetStreamContext{}
			mockJS := &sdk.NATSClient{}

			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mockResponse := `{"status": "ok"}`
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, mockResponse)
			}))
			defer testServer.Close()

			mockConfig := &config.KHConfig{
				PollInterval:    time.Second * 1,
				KuberhealthyURL: testServer.URL,
			}
			mockosexit := gomonkey.ApplyFunc(os.Exit, func(int) {

			})
			defer mockosexit.Reset()

			patchkhConfig := gomonkey.ApplyFunc(
				config.GetKuberHealthyConfig,
				func() (*config.KHConfig, error) {
					log.Println("before tt.wantErr####", tt.wantErr)
					if tt.wantErr {
						log.Println(tt.name, "******")
						return nil, errors.New("err")
					}
					return mockConfig, nil
				},
			)
			defer patchkhConfig.Reset()

			mockTicker := time.NewTicker(time.Second * 10)
			patchticker := gomonkey.ApplyFunc(
				time.NewTicker,
				func(time.Duration) *time.Ticker {
					return mockTicker
				},
			)
			defer patchticker.Reset()

			patchPollAndPublishKuberhealthy := gomonkey.ApplyFunc(
				mockPollAndPublishKuberhealthy,
				func(string, nats.JetStreamContext) error {
					if tt.publishErr {
						return errors.New("err")
					}
					return nil
				},
			)
			defer patchPollAndPublishKuberhealthy.Reset()

			patchPublishKuberhealthyMetrics := gomonkey.ApplyFunc(
				PublishKuberhealthyMetrics,
				func(js *sdk.NATSClient, state health.State) error {
					return nil
				},
			)
			defer patchPublishKuberhealthyMetrics.Reset()

			wg := sync.WaitGroup{}
			wg.Add(1)

			go func() {
				defer wg.Done()
				StartKuberHealthy(mockJS)
			}()
			if tt.wantErr {
				t.Fatalf("Error getting Kuberhealthy config1: err")

			}
		})
	}

}

func TestPollAndPublishKuberhealthy(t *testing.T) {
	cases := []struct {
		name         string
		wantErr      bool
		getErr       bool
		unMarshalErr bool
	}{
		{"success", false, false, false},
		{"failure readAll", true, false, false},
		{"failure unMarshal", false, false, true},
	}

	for _, tt := range cases {
		log.Println("Running test case: ", tt.name)
		mockJS := &sdk.NATSClient{}

		mockPublish := gomonkey.ApplyMethod(
			reflect.TypeOf(mockJS),
			"Publish",
			func(*sdk.NATSClient, string, []uint8) error {
				if tt.name == "error" {
					return errors.New("Error in publish")
				}
				return nil
			},
		)
		defer mockPublish.Reset()

		t.Run(tt.name, func(t *testing.T) {
			//mockJS := &MockJetStreamContext{}
			mockJS := &sdk.NATSClient{}

			if tt.wantErr {
				patchReadAll := gomonkey.ApplyFunc(
					io.ReadAll,
					func(io.Reader) ([]byte, error) {
						return nil, errors.New("err")
					},
				)
				defer patchReadAll.Reset()
			}
			if tt.getErr {
				patchGet := gomonkey.ApplyFunc(
					http.Get,
					func(string) (*http.Response, error) {
						return nil, errors.New("err")
					},
				)
				defer patchGet.Reset()
			}

			if tt.unMarshalErr {
				patchUnmarshal := gomonkey.ApplyFunc(
					json.Unmarshal,
					func([]byte, interface{}) error {
						return errors.New("err")
					},
				)
				defer patchUnmarshal.Reset()
			}

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mockState := health.State{
					OK: true,
				}
				mockStateJSON, _ := json.Marshal(mockState)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(mockStateJSON)
			}))
			defer ts.Close()

			err := pollAndPublishKuberhealthy(ts.URL, mockJS)
			if tt.wantErr {
				assert.Error(t, err)
			} else if tt.getErr {
				assert.Error(t, err)
			} else if tt.unMarshalErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

var mockhealthstate health.State

func TestPublishKuberhealthyMetrics(t *testing.T) {

	//mockJS := &MockJetStreamContext{}
	mockJS := &sdk.NATSClient{}

	dummyWorkloadDetails := khstatev1.WorkloadDetails{
		OK:               true,
		Errors:           []string{"No errors"},
		RunDuration:      "1s",
		Namespace:        "default",
		Node:             "node1",
		LastRun:          &metav1.Time{Time: time.Now()},
		AuthoritativePod: "pod1",
		CurrentUUID:      "12345678-1234-1234-1234-123456789abc",
	}
	mockResource := health.State{
		OK: true,
		CheckDetails: map[string]khstatev1.WorkloadDetails{
			"dummyCheck": dummyWorkloadDetails,
		},
		JobDetails: map[string]khstatev1.WorkloadDetails{
			"dummyJob": dummyWorkloadDetails,
		},
		CurrentMaster: "master1",
	}

	tests := []struct {
		name     string
		resource health.State
	}{
		{"success", mockResource},
		{"error", health.State{
			OK: false,
			CheckDetails: map[string]khstatev1.WorkloadDetails{
				"dummyCheck": dummyWorkloadDetails,
			},
			JobDetails: map[string]khstatev1.WorkloadDetails{
				"dummyJob": dummyWorkloadDetails,
			},
			CurrentMaster: "master1",
		}},
	}

	for i, tt := range tests {
		fmt.Println("Running test : ", i)

		mockPublish := gomonkey.ApplyMethod(
			reflect.TypeOf(mockJS),
			"Publish",
			func(*sdk.NATSClient, string, []uint8) error {
				if tt.name == "error" {
					return errors.New("Error in publish")
				}
				return nil
			},
		)
		defer mockPublish.Reset()

		t.Run(tt.name, func(t *testing.T) {

			err := PublishKuberhealthyMetrics(mockJS, tt.resource)
			if tt.name == "success" {

				assert.NoError(t, err)

				assert.NoError(t, err, "PublishKuberhealthyMetrics() returned an error")

				assert.Equal(t, mockResource.CurrentMaster, "master1", "CurrentMaster doesn't match")
			} else {
				fmt.Println("Error in Publish error: ", err)
				assert.EqualError(t, err, "Error in publish")
			}
		})
	}

	PublishKuberhealthyMetrics(mockJS, mockhealthstate)
}

type MockJetStreamContext struct{}

func (m *MockJetStreamContext) AccountInfo(opts ...nats.JSOpt) (*nats.AccountInfo, error) {
	return nil, nil
}

func (m *MockJetStreamContext) AddConsumer(stream string, cfg *nats.ConsumerConfig, opts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	return nil, nil
}
func (m *MockJetStreamContext) AddStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	return nil, nil
}

func (m *MockJetStreamContext) ChanQueueSubscribe(subj, queue string, ch chan *nats.Msg, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return nil, nil
}

func (m *MockJetStreamContext) ChanSubscribe(subj string, ch chan *nats.Msg, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return nil, nil
}

func (m *MockJetStreamContext) ConsumerInfo(stream, consumer string, opts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	return nil, nil
}
func (m *MockJetStreamContext) ConsumerNames(stream string, opts ...nats.JSOpt) <-chan string {
	return nil
}

func (m *MockJetStreamContext) Consumers(stream string, opts ...nats.JSOpt) <-chan *nats.ConsumerInfo {
	return nil
}

func (m *MockJetStreamContext) ConsumersInfo(stream string, opts ...nats.JSOpt) <-chan *nats.ConsumerInfo {
	return nil
}

func (m *MockJetStreamContext) CreateKeyValue(kv *nats.KeyValueConfig) (nats.KeyValue, error) {
	return nil, nil
}

func (m *MockJetStreamContext) CreateObjectStore(store *nats.ObjectStoreConfig) (nats.ObjectStore, error) {
	return nil, nil
}
func (m *MockJetStreamContext) DeleteConsumer(stream, consumer string, opts ...nats.JSOpt) error {
	return nil
}

func (m *MockJetStreamContext) DeleteKeyValue(key string) error {
	return nil
}

func (m *MockJetStreamContext) DeleteMsg(stream string, seq uint64, opts ...nats.JSOpt) error {
	return nil
}
func (m *MockJetStreamContext) DeleteObjectStore(store string) error {
	return nil
}

func (m *MockJetStreamContext) DeleteStream(stream string, opts ...nats.JSOpt) error {
	return nil
}

func (m *MockJetStreamContext) GetLastMsg(stream string, lastBy string, opts ...nats.JSOpt) (*nats.RawStreamMsg, error) {
	return nil, nil
}
func (m *MockJetStreamContext) GetMsg(stream string, seq uint64, opts ...nats.JSOpt) (*nats.RawStreamMsg, error) {
	return nil, nil
}

func (m *MockJetStreamContext) KeyValue(stream string) (nats.KeyValue, error) {
	return nil, nil
}

func (m *MockJetStreamContext) KeyValueStoreNames() <-chan string {
	return nil
}
func (m *MockJetStreamContext) KeyValueStores() <-chan nats.KeyValueStatus {
	return nil
}

func (m *MockJetStreamContext) ObjectStore(stream string) (nats.ObjectStore, error) {
	return nil, nil
}

func (m *MockJetStreamContext) ObjectStoreNames(opts ...nats.ObjectOpt) <-chan string {
	return nil
}

func (m *MockJetStreamContext) ObjectStores(opts ...nats.ObjectOpt) <-chan nats.ObjectStoreStatus {
	return nil
}

func (m *MockJetStreamContext) PublishAsync(subj string, data []byte, opts ...nats.PubOpt) (nats.PubAckFuture, error) {
	return nil, nil
}

func (m *MockJetStreamContext) PublishAsyncComplete() <-chan struct{} {
	return nil
}
func (m *MockJetStreamContext) PublishAsyncPending() int {
	return 0
}

func (m *MockJetStreamContext) PublishMsg(msg *nats.Msg, opts ...nats.PubOpt) (*nats.PubAck, error) {
	return nil, nil
}

func (m *MockJetStreamContext) PublishMsgAsync(msg *nats.Msg, opts ...nats.PubOpt) (nats.PubAckFuture, error) {
	return nil, nil
}
func (m *MockJetStreamContext) PullSubscribe(subject, queue string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return nil, nil
}

func (m *MockJetStreamContext) PurgeStream(stream string, opts ...nats.JSOpt) error {
	return nil
}

func (m *MockJetStreamContext) QueueSubscribe(subject, queue string, handler nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return nil, nil
}

func (m *MockJetStreamContext) Publish(subj string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error) {
	resource := &health.State{}
	json.Unmarshal(data, resource)
	if resource.OK == false {
		return nil, errors.New("Error in publish")
	}
	return nil, nil
}

func (m *MockJetStreamContext) QueueSubscribeSync(subject, queue string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return nil, nil
}

func (m *MockJetStreamContext) SecureDeleteMsg(stream string, seq uint64, opts ...nats.JSOpt) error {
	return nil
}

func (m *MockJetStreamContext) StreamInfo(stream string, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	return nil, nil
}

func (m *MockJetStreamContext) StreamNameBySubject(subject string, opts ...nats.JSOpt) (string, error) {
	return "", nil
}

func (m *MockJetStreamContext) StreamNames(opts ...nats.JSOpt) <-chan string {
	return nil
}

func (m *MockJetStreamContext) Streams(opts ...nats.JSOpt) <-chan *nats.StreamInfo {
	return nil
}

func (m *MockJetStreamContext) StreamsInfo(opts ...nats.JSOpt) <-chan *nats.StreamInfo {
	return nil
}

func (m *MockJetStreamContext) Subscribe(subject string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return nil, nil
}

func (m *MockJetStreamContext) SubscribeSync(subject string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return nil, nil
}

func (m *MockJetStreamContext) UpdateConsumer(stream string, cfg *nats.ConsumerConfig, opts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	return nil, nil
}

func (m *MockJetStreamContext) UpdateStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	return nil, nil
}

func mockPollAndPublishKuberhealthy(url string, js nats.JetStreamContext) error {
	return nil
}
