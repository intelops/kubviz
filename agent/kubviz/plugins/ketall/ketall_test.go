package ketall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Resource struct {
	Resource    string `json:"resource"`
	Namespace   string `json:"namespace"`
	ClusterName string `json:"clusterName"`
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

func (m *MockJetStreamContext) Publish(subj string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error) {
	resource := &Resource{}
	json.Unmarshal(data, resource)
	if resource.Resource == "test-error" {
		return nil, errors.New("Error in publish")
	}
	return nil, nil
}

type MockResourceInterface struct{}

func (m *MockResourceInterface) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}
func (m *MockResourceInterface) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}
func (m *MockResourceInterface) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}
func (m *MockResourceInterface) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	return nil
}
func (m *MockResourceInterface) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return nil
}
func (m *MockResourceInterface) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *MockResourceInterface) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	fmt.Println("List 1 called")
	return &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "List",
		},
	}, nil
}

func (m *MockResourceInterface) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}
func (m *MockResourceInterface) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}
func (m *MockResourceInterface) Apply(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}
func (m *MockResourceInterface) ApplyStatus(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}

type MockNamespaceableResourceInterface struct{}

func (m *MockNamespaceableResourceInterface) Namespace(s string) dynamic.ResourceInterface {
	return &MockResourceInterface{}
}

func TestPublishAllResources(t *testing.T) {
	mockResource := model.Resource{
		Resource:    "test-resource",
		Kind:        "test-kind",
		Namespace:   "test-namespace",
		Age:         "test-age",
		ClusterName: "test-cluster",
	}
	tests := []struct {
		name     string
		resource model.Resource
	}{
		{"success", mockResource},
		{"error", model.Resource{}},
	}
	for _, tt := range tests {
		mockJS := &MockJetStreamContext{}

		mockPublish := gomonkey.ApplyMethod(
			reflect.TypeOf(mockJS),
			"Publish",
			func(*MockJetStreamContext, string, []byte, ...nats.PubOpt) (*nats.PubAck, error) {
				if tt.name == "error" {
					return nil, errors.New("Error in publish")
				}
				return nil, nil
			},
		)
		defer mockPublish.Reset()

		t.Run(tt.name, func(t *testing.T) {

			err := PublishAllResources(tt.resource, mockJS)
			if tt.name == "error" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type MockDynamicResource struct {
	FnList      func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error)
	FnApply     func(ctx context.Context, namespace string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error)
	FnApplyStat func(ctx context.Context, namespace string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error)
}

func (m *MockDynamicResource) List(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
	fmt.Println("List 2 called")
	return m.FnList(ctx, options)
}

func (m *MockDynamicResource) Apply(ctx context.Context, namespace string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return m.FnApply(ctx, namespace, obj, options, subresources...)
}

func (m *MockDynamicResource) ApplyStatus(ctx context.Context, namespace string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return m.FnApplyStat(ctx, namespace, obj, options, subresources...)
}

func (m *MockNamespaceableResourceInterface) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *MockNamespaceableResourceInterface) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *MockNamespaceableResourceInterface) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *MockNamespaceableResourceInterface) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	return nil
}

func (m *MockNamespaceableResourceInterface) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return nil
}

func (m *MockNamespaceableResourceInterface) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *MockNamespaceableResourceInterface) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	fmt.Println("List 3 called")
	return nil, nil
}

func (m *MockNamespaceableResourceInterface) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (m *MockNamespaceableResourceInterface) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *MockNamespaceableResourceInterface) Apply(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *MockNamespaceableResourceInterface) ApplyStatus(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}

type APIPathResolverFunc func(kind schema.GroupVersionKind) string

func LegacyAPIPathResolverFunc(kind schema.GroupVersionKind) string {
	if len(kind.Group) == 0 {
		return "/api"
	}
	return "/apis"
}

func TestGetAllResources2(t *testing.T) {
	cases := []struct {
		name                   string
		isNameSpaceEmpty       bool
		wantErr                bool
		PublishAllResourcesErr bool
	}{
		{"success with namespace", false, false, false},
		{"success without namespace", true, false, false},
		{"error in NewForConfig", false, true, false},
	}

	for _, tt := range cases {
		mockConfig := &rest.Config{}
		mockJS := &MockJetStreamContext{}
		mockDC := &discovery.DiscoveryClient{}
		mockGroupVersionResource := schema.GroupVersionResource{
			Group:    "group",
			Version:  "version",
			Resource: "resource",
		}
		mockgvrs := make(map[schema.GroupVersionResource]struct{})
		mockgvrs[mockGroupVersionResource] = struct{}{}
		mockList := &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "group/version",
						"kind":       "resource",
					},
				},
			},
		}
		mockDyC := &dynamic.DynamicClient{}
		mockNamespace := &MockNamespaceableResourceInterface{}
		mockResourceInterface := &MockResourceInterface{}

		patchNewDiscovery := gomonkey.ApplyFunc(
			discovery.NewDiscoveryClientForConfigOrDie,
			func(*rest.Config) *discovery.DiscoveryClient {
				return mockDC
			},
		)
		defer patchNewDiscovery.Reset()

		if tt.wantErr {
			patchNewDynamic := gomonkey.ApplyFunc(
				dynamic.NewForConfig,
				func(*rest.Config) (*dynamic.DynamicClient, error) {
					return nil, errors.New("new dynamic client error")
				},
			)
			defer patchNewDynamic.Reset()
		}

		patchResourceLists := gomonkey.ApplyMethod(
			reflect.TypeOf(mockDC),
			"ServerPreferredResources",
			func(*discovery.DiscoveryClient) ([]*metav1.APIResourceList, error) {
				return []*metav1.APIResourceList{}, nil
			},
		)
		defer patchResourceLists.Reset()

		patchgvrs := gomonkey.ApplyFunc(
			discovery.GroupVersionResources,
			func([]*metav1.APIResourceList) (map[schema.GroupVersionResource]struct{}, error) {
				return mockgvrs, nil
			},
		)
		defer patchgvrs.Reset()

		mockDynamicResourceInterface := &MockNamespaceableResourceInterface{}
		patchResource := gomonkey.ApplyMethod(
			reflect.TypeOf(mockDyC),
			"Resource",
			func(*dynamic.DynamicClient, schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
				return mockDynamicResourceInterface
			},
		)
		defer patchResource.Reset()

		patchNamespace := gomonkey.ApplyMethod(
			reflect.TypeOf(mockNamespace),
			"Namespace",
			func(*MockNamespaceableResourceInterface, string) dynamic.ResourceInterface {
				return mockResourceInterface
			},
		)
		defer patchNamespace.Reset()

		patchList := gomonkey.ApplyMethod(
			reflect.TypeOf(mockResourceInterface),
			"List",
			func(*MockResourceInterface, context.Context, metav1.ListOptions) (*unstructured.UnstructuredList, error) {
				return mockList, nil
			},
		)
		defer patchList.Reset()

		mockItem := &unstructured.Unstructured{}
		patchGetNamespace := gomonkey.ApplyMethod(
			reflect.TypeOf(mockItem),
			"GetNamespace",
			func(*unstructured.Unstructured) string {
				if tt.isNameSpaceEmpty {
					return ""
				}
				return "default"
			},
		)
		defer patchGetNamespace.Reset()

		t.Run(tt.name, func(t *testing.T) {
			err := GetAllResources(mockConfig, mockJS)
			fmt.Println("Error in GetAllResources: ", err)
			if tt.wantErr {
				require.Error(t, err)
			} else if tt.isNameSpaceEmpty {
				require.NoError(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
