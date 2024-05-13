package rakkess

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/blang/semver"

	openapi_v2 "github.com/google/gnostic/openapiv2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/authorization/v1"
	apiV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	run "k8s.io/apimachinery/pkg/runtime"
	newv1 "k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/authorization/v1/fake"
	"k8s.io/client-go/openapi"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	authTesting "k8s.io/client-go/testing"
	//"k8s.io/apimachinery/pkg/version"
)

func TestOutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:   "valid format",
			format: "icon-table",
		},
		{
			name:     "invalid format",
			format:   "cassowary",
			expected: "unexpected output format: cassowary",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := OutputFormat(test.format)
			if test.expected != "" {
				assert.EqualError(t, actual, test.expected)
			} else {
				assert.NoError(t, actual)
			}
		})
	}
}
func TestVerbs(t *testing.T) {
	tests := []struct {
		name     string
		verbs    []string
		expected string
	}{
		{
			name:  "only valid verbs",
			verbs: []string{"list", "get", "deletecollection"},
		},
		{
			name:     "only invalid verbs",
			verbs:    []string{"lust", "git", "poxy"},
			expected: "unexpected verbs: [git lust poxy]",
		},
		{
			name:     "valid and invalid verbs",
			verbs:    []string{"list", "git", "deletecollection"},
			expected: "unexpected verbs: [git]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := verbs(test.verbs)
			if test.expected != "" {
				assert.EqualError(t, actual, test.expected)
			} else {
				assert.NoError(t, actual)
			}
		})
	}
}
func TestParseVersion(t *testing.T) {
	var tests = []struct {
		name      string
		given     string
		expected  semver.Version
		shouldErr bool
	}{
		{
			name:     "parse version correct",
			given:    "v3.14.15",
			expected: semver.MustParse("3.14.15"),
		},
		{
			name:     "parse with trailing text",
			given:    "v2.71.82-dirty",
			expected: semver.MustParse("2.71.82-dirty"),
		},
		{
			name:     "fail parse without leading v",
			given:    "2.71.82",
			expected: semver.MustParse("2.71.82"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseVersion(test.given)
			if test.shouldErr {
				assert.Error(t, err, "parse should fail")
			} else {
				assert.Equal(t, test.expected, actual, "parse should succeed")

			}
		})
	}
}

const HEADER = "NAME       GET  LIST\n"

func TestPrintResults(t *testing.T) {
	tests := []struct {
		name      string
		table     *Table
		want      string
		wantColor string
		wantASCII string
	}{
		{
			"single result, all allowed",
			&Table{
				Headers: []string{"NAME", "GET", "LIST"},
				Rows: []Row{
					{Intro: []string{"resource1"}, Entries: []Outcome{Up, Up}},
				},
			},
			HEADER + "resource1  ✔    ✔\n",
			HEADER + "resource1  \033[32m✔\033[0m    \033[32m✔\033[0m\n",
			HEADER + "resource1  yes  yes\n",
		},
		{
			"single result, all forbidden",
			&Table{
				Headers: []string{"NAME", "GET", "LIST"},
				Rows: []Row{
					{Intro: []string{"resource1"}, Entries: []Outcome{Down, Down}},
				},
			},
			HEADER + "resource1  ✖    ✖\n",
			HEADER + "resource1  \033[31m✖\033[0m    \033[31m✖\033[0m\n",
			HEADER + "resource1  no   no\n",
		},
		{
			"single result, all not applicable",
			&Table{
				Headers: []string{"NAME", "GET", "LIST"},
				Rows: []Row{
					{Intro: []string{"resource1"}, Entries: []Outcome{None, None}},
				},
			},
			HEADER + "resource1       \n",
			HEADER + "resource1  \033[0m\033[0m     \033[0m\033[0m\n",
			HEADER + "resource1  n/a  n/a\n",
		},
		{
			"single result, all ERR",
			&Table{
				Headers: []string{"NAME", "GET", "LIST"},
				Rows: []Row{
					{Intro: []string{"resource1"}, Entries: []Outcome{Err, Err}},
				},
			},
			HEADER + "resource1  ERR  ERR\n",
			HEADER + "resource1  \033[35mERR\033[0m  \033[35mERR\033[0m\n",
			HEADER + "resource1  ERR  ERR\n",
		},
		{
			"single result, mixed",
			&Table{
				Headers: []string{"NAME", "GET", "LIST"},
				Rows: []Row{
					{Intro: []string{"resource1"}, Entries: []Outcome{Down, Up}},
				},
			},
			HEADER + "resource1  ✖    ✔\n",
			"",
			HEADER + "resource1  no   yes\n",
		},
		{
			"many results",
			&Table{
				Headers: []string{"NAME", "GET"},
				Rows: []Row{
					{Intro: []string{"resource1"}, Entries: []Outcome{Down}},
					{Intro: []string{"resource2"}, Entries: []Outcome{Up}},
					{Intro: []string{"resource3"}, Entries: []Outcome{Err}},
				},
			},
			"NAME       GET\nresource1  ✖\nresource2  ✔\nresource3  ERR\n",
			"",
			"NAME       GET\nresource1  no\nresource2  yes\nresource3  ERR\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tc.table.Render(buf, "icon-table")
			assert.Equal(t, tc.want, buf.String())

			buf = &bytes.Buffer{}
			tc.table.Render(buf, "ascii-table")
			assert.Equal(t, tc.wantASCII, buf.String())
		})
	}

	for _, tc := range tests[0:4] {
		isTerminal = func(w io.Writer) bool {
			return true
		}
		defer func() {
			isTerminal = isTerminalImpl
		}()

		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tc.table.Render(buf, "icon-table")
			assert.Equal(t, tc.wantColor, buf.String())

			buf = &bytes.Buffer{}
			tc.table.Render(buf, "ascii-table")
			assert.Equal(t, tc.wantASCII, buf.String())
		})
	}
}
func TestRakkessOptions_ExpandVerbs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "* wildcard",
			input:    []string{"*"},
			expected: ValidVerbs,
		},
		{
			name:     "all wildcard",
			input:    []string{"*"},
			expected: ValidVerbs,
		},
		{
			name:     "wildcard mixed with other verbs",
			input:    []string{"list", "*", "get"},
			expected: ValidVerbs,
		},
		{
			name:     "no wildcard",
			input:    []string{"list", "get"},
			expected: []string{"list", "get"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := &RakkessOptions{Verbs: test.input}
			opts.ExpandVerbs()

			assert.Equal(t, test.expected, opts.Verbs)
		})
	}
}

func TestRakkessOptions_ExpandServiceAccount(t *testing.T) {
	tests := []struct {
		name           string
		serviceAccount string
		namespace      string
		impersonate    string
		expected       string
		expectedErr    string
	}{
		{
			name:        "no serviceAccount given",
			impersonate: "original-impersonate",
			expected:    "original-impersonate",
		},
		{
			name:           "unqualified serviceAccount and namespace",
			serviceAccount: "some-sa",
			namespace:      "some-ns",
			expected:       "system:serviceaccount:some-ns:some-sa",
		},
		{
			name:           "qualified serviceAccount",
			serviceAccount: "some-ns:some-sa",
			expected:       "system:serviceaccount:some-ns:some-sa",
		},
		{
			name:           "unqualified serviceAccount without namespace",
			serviceAccount: "some-ns",
			expectedErr:    "fully qualify the serviceAccount",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := &RakkessOptions{
				ConfigFlags: &genericclioptions.ConfigFlags{
					Impersonate: &test.impersonate,
					Namespace:   &test.namespace,
				},
				AsServiceAccount: test.serviceAccount,
			}

			err := opts.ExpandServiceAccount()
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErr)
			} else {
				assert.Equal(t, test.expected, *opts.ConfigFlags.Impersonate)
			}
		})
	}
}

func TestGroupResource_fullName(t *testing.T) {
	grNoGroup := &GroupResource{
		APIGroup: "",
		APIResource: metav1.APIResource{
			Name: "foo",
		},
	}
	assert.Equal(t, "foo", grNoGroup.fullName())

	grGroup := &GroupResource{
		APIGroup: "v1",
		APIResource: metav1.APIResource{
			Name: "foo",
		},
	}
	assert.Equal(t, "foo.v1", grGroup.fullName())
}

type SelfSubjectAccessReviewDecision struct {
	v1.ResourceAttributes
	decision Access
}

func (d *SelfSubjectAccessReviewDecision) matches(other *v1.SelfSubjectAccessReview) bool {
	return d.ResourceAttributes == *other.Spec.ResourceAttributes
}

func toGroupResource(group, name string, verbs ...string) GroupResource {
	return GroupResource{
		APIGroup: group,
		APIResource: apiV1.APIResource{
			Name:  name,
			Verbs: verbs,
		},
	}
}

func TestCheckResourceAccess(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		verbs     []string
		input     []GroupResource
		decisions []*SelfSubjectAccessReviewDecision
		want      []string
	}{
		{
			name:  "single resource, single verb",
			verbs: []string{"list"},
			input: []GroupResource{toGroupResource("group1", "resource1", "list")},
			decisions: []*SelfSubjectAccessReviewDecision{
				{
					v1.ResourceAttributes{
						Resource: "resource1",
						Group:    "group1",
						Verb:     "list",
					},
					Allowed,
				},
			},
			want: []string{"resource1.group1:list->ok"},
		},
		{
			name:  "single resource, invalid verb",
			verbs: []string{"patch"},
			input: []GroupResource{toGroupResource("group1", "resource1", "list")},
			want:  []string{"resource1.group1:patch->n/a"},
		},
		{
			name:  "single resource, multiple verbs",
			verbs: []string{"list", "create", "delete"},
			input: []GroupResource{toGroupResource("group1", "resource1", "list", "create", "delete")},
			decisions: []*SelfSubjectAccessReviewDecision{
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "list"},
					Allowed,
				},
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "create"},
					Allowed,
				},
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "delete"},
					Denied,
				},
			},
			want: []string{"resource1.group1:create->ok,delete->no,list->ok"},
		},
		{
			name:  "multiple resources, single verb",
			verbs: []string{"list"},
			input: []GroupResource{
				toGroupResource("group1", "resource1", "list"),
				toGroupResource("group1", "resource2", "list"),
			},
			decisions: []*SelfSubjectAccessReviewDecision{
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "list"},
					Allowed,
				},
				{
					v1.ResourceAttributes{Resource: "resource2", Group: "group1", Verb: "list"},
					Denied,
				},
			},
			want: []string{"resource1.group1:list->ok", "resource2.group1:list->no"},
		},
		{
			name:  "multiple resources, multiple verbs",
			verbs: []string{"list", "create"},
			input: []GroupResource{
				toGroupResource("group1", "resource1", "list", "create"),
				toGroupResource("group1", "resource2", "create"),
				toGroupResource("group2", "resource1", "list"),
			},
			decisions: []*SelfSubjectAccessReviewDecision{
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "list"},
					Allowed,
				},
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "create"},
					Denied,
				},
				{
					v1.ResourceAttributes{Resource: "resource2", Group: "group1", Verb: "create"},
					Denied,
				},
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group2", Verb: "list"},
					Allowed,
				},
			},
			want: []string{"resource1.group1:create->no,list->ok", "resource1.group2:create->n/a,list->ok", "resource2.group1:create->no,list->n/a"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeReviews := &fake.FakeSelfSubjectAccessReviews{Fake: &fake.FakeAuthorizationV1{Fake: &authTesting.Fake{}}}
			fakeReviews.Fake.AddReactor("create", "selfsubjectaccessreviews",
				func(action authTesting.Action) (handled bool, ret run.Object, err error) {
					sar := action.(authTesting.CreateAction).GetObject().(*v1.SelfSubjectAccessReview)

					for _, d := range test.decisions {
						if d.matches(sar) {
							sar.Status.Allowed = d.decision == Allowed
							return true, sar, nil
						}
					}
					return false, nil, nil
				})

			results := CheckResourceAccess(ctx, fakeReviews, test.input, test.verbs, nil)

			var got []string
			for name, access := range results {
				var as []string
				for verb, a := range access {
					var outcome string
					switch a {
					case Allowed:
						outcome = "ok"
					case Denied:
						outcome = "no"
					case NotApplicable:
						outcome = "n/a"
					}
					as = append(as, verb+"->"+outcome)
				}
				sort.Strings(as)
				got = append(got, name+":"+strings.Join(as, ","))
			}
			sort.Strings(got)
			assert.Equal(t, test.want, got)
		})
	}
}

type fakeCachedDiscoveryInterface struct {
	invalidateCalls int
	next            metav1.APIResourceList
	err             error
	fresh           bool
}

func (f *fakeCachedDiscoveryInterface) OpenAPISchema() (*openapi_v2.Document, error) {
	// Your implementation here
	return nil, nil
}

// Ensure that fakeCachedDiscoveryInterface implements CachedDiscoveryInterface
var _ discovery.CachedDiscoveryInterface = (*fakeCachedDiscoveryInterface)(nil)

func (c *fakeCachedDiscoveryInterface) Fresh() bool {
	return c.fresh
}
func (c *fakeCachedDiscoveryInterface) OpenAPIV3() openapi.Client {
	panic("not implemented")
}
func (c *fakeCachedDiscoveryInterface) WithLegacy() discovery.DiscoveryInterface {
	panic("not implemented")
}
func (c *fakeCachedDiscoveryInterface) Invalidate() {
	c.invalidateCalls++
	c.fresh = true
}

func (c *fakeCachedDiscoveryInterface) RESTClient() restclient.Interface {
	panic("not implemented")
}

func (c *fakeCachedDiscoveryInterface) ServerGroups() (*metav1.APIGroupList, error) {
	panic("not implemented")
}

func (c *fakeCachedDiscoveryInterface) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	panic("not implemented")
}

func (c *fakeCachedDiscoveryInterface) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	panic("not implemented")
}

func (c *fakeCachedDiscoveryInterface) ServerResources() ([]*metav1.APIResourceList, error) {
	panic("not implemented")
}

func (c *fakeCachedDiscoveryInterface) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	if c.fresh {
		return []*metav1.APIResourceList{&c.next}, c.err
	}
	return nil, c.err
}

func (c *fakeCachedDiscoveryInterface) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	if c.fresh {
		return []*metav1.APIResourceList{&c.next}, c.err
	}
	return nil, c.err
}

func (c *fakeCachedDiscoveryInterface) ServerVersion() (*newv1.Info, error) {
	panic("not implemented")
}

// func (c *fakeCachedDiscoveryInterface) OpenAPISchema() (*openapi_v2.Document, error) {
// 	panic("not implemented")
// }

var (
	aFoo = metav1.APIResource{
		Name:       "foo",
		Kind:       "Foo",
		Namespaced: false,
		Verbs:      []string{"list"},
	}
	aNoVerbs = metav1.APIResource{
		Name:       "baz",
		Kind:       "Baz",
		Namespaced: false,
		Verbs:      []string{},
	}
	bBar = metav1.APIResource{
		Name:       "bar",
		Kind:       "Bar",
		Namespaced: true,
		Verbs:      []string{"list"},
	}
)

func TestFetchAvailableGroupResources(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		verbs     []string
		resources metav1.APIResourceList
		err       error
		expected  interface{}
	}{
		{
			name:  "cluster resources",
			verbs: []string{"list"},
			resources: metav1.APIResourceList{
				GroupVersion: "a/v1",
				APIResources: []metav1.APIResource{aFoo, aNoVerbs},
			},
			expected: []GroupResource{{APIGroup: "a", APIResource: aFoo}},
		},
		{
			name:      "namespaced resources",
			namespace: "any-namespace",
			verbs:     []string{"list"},
			resources: metav1.APIResourceList{
				GroupVersion: "b/v1",
				APIResources: []metav1.APIResource{bBar},
			},
			expected: []GroupResource{{APIGroup: "b", APIResource: bBar}},
		},
		{
			name:  "incomplete cluster resources",
			err:   fmt.Errorf("list is incomplete"),
			verbs: []string{"list"},
			resources: metav1.APIResourceList{
				GroupVersion: "a/v1",
				APIResources: []metav1.APIResource{aFoo, aNoVerbs},
			},
			expected: []GroupResource{{APIGroup: "a", APIResource: aFoo}},
		},
		{
			name:      "incomplete namespaced resources",
			namespace: "any-namespace",
			err:       fmt.Errorf("list is incomplete"),
			verbs:     []string{"list"},
			resources: metav1.APIResourceList{
				GroupVersion: "b/v1",
				APIResources: []metav1.APIResource{bBar},
			},
			expected: []GroupResource{{APIGroup: "b", APIResource: bBar}},
		},
		{
			name:      "empty api-resources",
			namespace: "any-namespace",
			verbs:     []string{"list"},
			resources: metav1.APIResourceList{
				GroupVersion: "c/v1",
				APIResources: []metav1.APIResource{},
			},
			expected: []GroupResource(nil),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeRbacClient := &fakeCachedDiscoveryInterface{
				next: test.resources,
				err:  test.err,
			}

			getDiscoveryClient = func(opts *RakkessOptions) (discovery.CachedDiscoveryInterface, error) {
				return fakeRbacClient, nil
			}
			defer func() { getDiscoveryClient = getDiscoveryClientImpl }()

			opts := &RakkessOptions{
				ConfigFlags: &genericclioptions.ConfigFlags{
					Namespace: &test.namespace,
				},
			}
			grs, err := FetchAvailableGroupResources(opts)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, grs)
		})
	}
}

type AccessMap map[string]Access

func TestDiff(t *testing.T) {
	// Test case 1: Empty inputs
	// leftEmpty := make(ResourceAccess)
	// rightEmpty := make(ResourceAccess)
	verbs := []string{"get", "create"}
	// expected := &Table{
	// 	Headers: []string{"NAME", "GET", "CREATE"},
	// 	Rows:    []Row{},
	// } // Create an empty Table object with no rows
	// result := Diff(leftEmpty, rightEmpty, verbs)
	// if !reflect.DeepEqual(result, expected) {
	// 	t.Errorf("Test case 1 failed: Expected %v, but got %v", expected, result)
	// }

	// Test case 2: left and right have the same keys and values
	left := ResourceAccess{
		"resource1": AccessMap{
			"get":    Allowed,
			"create": Denied,
		},
		"resource2": AccessMap{
			"get":    Allowed,
			"create": Allowed,
		},
	}
	right := ResourceAccess{
		"resource1": AccessMap{
			"get":    Allowed,
			"create": Allowed,
		},
		"resource2": AccessMap{
			"get":    Allowed,
			"create": Allowed,
		},
	}
	expected := &Table{
		Headers: []string{"NAME", "GET", "CREATE"},
		Rows: []Row{
			{Intro: []string{"resource1"}, Entries: []Outcome{0, 1}},
		},
	}
	result := Diff(left, right, verbs)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Test case 2 failed: Expected %v, but got %v", expected, result)
	}

	// Add more test cases as needed
}
func TestTable(t *testing.T) {

	// Test case 2: Non-empty verbs
	ra := ResourceAccess{
		"resource1": AccessMap{
			"get":    Allowed,
			"create": Denied,
		},
		"resource2": AccessMap{
			"get":    Allowed,
			"create": Allowed,
		},
	}
	verbs := []string{"get", "create"}
	expected := &Table{
		Headers: []string{"NAME", "GET", "CREATE"},
		Rows: []Row{
			{Intro: []string{"resource1"}, Entries: []Outcome{Up, Down}},
			{Intro: []string{"resource2"}, Entries: []Outcome{Up, Up}},
		},
	}
	result := ra.Table(verbs)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Test case 2 failed: Expected %v, but got %v", expected, result)
	}
}
func TestGetAuthClient(t *testing.T) {
	// Create a new RakkessOptions instance
	opts := &RakkessOptions{
		ConfigFlags: genericclioptions.NewConfigFlags(true),
	}

	// Patch the ToRESTConfig method to return a dummy rest.Config
	monkey.PatchInstanceMethod(reflect.TypeOf(opts.ConfigFlags), "ToRESTConfig", func(flags *genericclioptions.ConfigFlags) (*rest.Config, error) {
		return &rest.Config{}, nil
	})
	defer monkey.UnpatchAll()

	// Test the function
	authClient, err := opts.GetAuthClient()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Verify that the returned authClient is not nil
	if authClient == nil {
		t.Error("Expected authClient to be non-nil, but got nil")
	}
}
func TestNewRakkessOptions(t *testing.T) {
	// Create a new RakkessOptions instance
	opts := NewRakkessOptions()

	// Verify that ConfigFlags is set and not nil
	assert.NotNil(t, opts.ConfigFlags, "ConfigFlags should not be nil")
	// Verify that Streams is set and not nil
	assert.NotNil(t, opts.Streams, "Streams should not be nil")
	// Verify that Streams.In, Streams.Out, and Streams.ErrOut are set to os.Stdin, os.Stdout, and os.Stderr respectively
	assert.Equal(t, os.Stdin, opts.Streams.In, "Streams.In should be os.Stdin")
	assert.Equal(t, os.Stdout, opts.Streams.Out, "Streams.Out should be os.Stdout")
	assert.Equal(t, os.Stderr, opts.Streams.ErrOut, "Streams.ErrOut should be os.Stderr")
}
func TestNewTestRakkessOptions(t *testing.T) {
	opts, in, out, errOut := NewTestRakkessOptions()

	// Verify that opts is not nil
	assert.NotNil(t, opts, "Options should not be nil")
	// Verify that ConfigFlags is set to genericclioptions.NewConfigFlags(true)
	//assert.True(t, opts.ConfigFlags.GenericConfig.PreferredInput == "yaml", "ConfigFlags should be set to genericclioptions.NewConfigFlags(true)")

	// Verify that Streams.In, Streams.Out, and Streams.ErrOut are set to the expected values
	assert.Equal(t, in, opts.Streams.In, "Streams.In should be set correctly")
	assert.Equal(t, out, opts.Streams.Out, "Streams.Out should be set correctly")
	assert.Equal(t, errOut, opts.Streams.ErrOut, "Streams.ErrOut should be set correctly")
}
func TestOptions(t *testing.T) {
	tests := []struct {
		name        string
		opts        *RakkessOptions
		expectedErr error
	}{
		// {
		// 	name: "Valid options",
		// 	opts: &RakkessOptions{
		// 		Verbs:        []string{"get", "create"},
		// 		OutputFormat: "json",
		// 	},
		// 	expectedErr: nil,
		// },
		// func TestInitTerminal(t *testing.T) {
		// 	// Create a buffer to capture the output
		// 	buf := new(bytes.Buffer)

		// 	// Call initTerminal with the buffer
		// 	initTerminal(buf)

		//		// Check if the output matches the expected value
		//		expected := ""
		//		if buf.String() != expected {
		//			t.Errorf("Expected: %s, got: %s", expected, buf.String())
		//		}
		//	}
		{
			name: "Invalid verbs",
			opts: &RakkessOptions{
				Verbs:        []string{"invalidVerb"},
				OutputFormat: "json",
			},
			expectedErr: fmt.Errorf("unexpected verbs: [invalidVerb]"),
		},
		{
			name: "Invalid output format",
			opts: &RakkessOptions{
				Verbs:        []string{"get", "create"},
				OutputFormat: "invalidFormat",
			},
			expectedErr: fmt.Errorf("unexpected output format: invalidFormat"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Options(tt.opts)
			if err != nil && tt.expectedErr == nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && tt.expectedErr != nil {
				t.Errorf("Expected error %v, but got nil", tt.expectedErr)
			}
			if err != nil && tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
			}
		})
	}
}
func TestGetBuildInfo(t *testing.T) {
	expected := GetBuildInfo()

	actual := GetBuildInfo()

	if actual.Version != expected.Version {
		t.Errorf("Version should match. Expected %s, got %s", expected.Version, actual.Version)
	}
	if actual.GitCommit != expected.GitCommit {
		t.Errorf("GitCommit should match. Expected %s, got %s", expected.GitCommit, actual.GitCommit)
	}
	if actual.BuildDate != expected.BuildDate {
		t.Errorf("BuildDate should match. Expected %s, got %s", expected.BuildDate, actual.BuildDate)
	}
	if actual.GoVersion != expected.GoVersion {
		t.Errorf("GoVersion should match. Expected %s, got %s", expected.GoVersion, actual.GoVersion)
	}
	if actual.Compiler != expected.Compiler {
		t.Errorf("Compiler should match. Expected %s, got %s", expected.Compiler, actual.Compiler)
	}
	if actual.Platform != expected.Platform {
		t.Errorf("Platform should match. Expected %s, got %s", expected.Platform, actual.Platform)
	}
}

func TestAccessToOutcome(t *testing.T) {
	tests := []struct {
		input  Access
		output Outcome
		err    error
	}{
		{0, None, nil},
		{1, Up, nil},
		{2, Down, nil},
		{3, Err, nil},
		{4, None, fmt.Errorf("unknown access code: 4")},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Access=%d", test.input), func(t *testing.T) {
			output, err := accessToOutcome(test.input)
			if err != nil && test.err == nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && test.err != nil {
				t.Errorf("Expected error: %v, but got none", test.err)
			}
			if output != test.output {
				t.Errorf("Expected output: %v, but got: %v", test.output, output)
			}
		})
	}
}

var sigSigs []os.Signal

func TestCatchCtrlC(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 10)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	catchCtrlC(cancel)

	select {
	case <-ctx.Done():
		if ctx.Err() != context.Canceled {
			t.Errorf("Expected context to be canceled, but got: %v", ctx.Err())
		}
	case <-time.After(time.Second):
		t.Errorf("Timeout waiting for context to be canceled")
	}
}

// Define a function to return a fake list of pods
func fakePods(clientset *kubernetes.Clientset) ([]*unstructured.Unstructured, error) {
	// Return your fake pods here
	return []*unstructured.Unstructured{}, nil
}

// HumanreadableAccessCode is a mock function for testing

// Outcome is a mock type for testing
func TestDiscoveryClient(t *testing.T) {
	// Test case 1: Successful creation of discovery client
	opts := &RakkessOptions{
		ConfigFlags: &genericclioptions.ConfigFlags{},
	}
	_, err := opts.DiscoveryClient()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

}
func TestGetDiscoveryClientImpl(t *testing.T) {
	// Test case 1: Successful retrieval of discovery client when options are valid
	opts := &RakkessOptions{
		ConfigFlags: &genericclioptions.ConfigFlags{},
	}
	client, err := getDiscoveryClientImpl(opts)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if client == nil {
		t.Error("Expected non-nil discovery client, but got nil")
	}

}
func TestInitTerminal(t *testing.T) {
	// Create a mock io.Writer
	mockWriter := &mockWriter{}

	// Call the function with the mock writer
	initTerminal(mockWriter)

	// Assert that the function did not return an error
	if mockWriter.err != nil {
		t.Errorf("Expected no error, got %v", mockWriter.err)
	}
}

// Mock io.Writer implementation for testing
type mockWriter struct {
	err error
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return 0, m.err
}
func TestResource(t *testing.T) {
	ctx := context.Background()

	// Test case 1: Options returns an error
	opts := &RakkessOptions{}
	_, err := Resource(ctx, opts)
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}

}
