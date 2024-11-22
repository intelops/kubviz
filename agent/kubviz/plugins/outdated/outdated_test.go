package outdated

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/genuinetools/reg/registry"
	"github.com/hashicorp/go-version"
	semver "github.com/hashicorp/go-version"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/nats/sdk"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func TestTruncateImageName(t *testing.T) {
	input := "docker.io/library/nginx:latest12345678901234567890123456789012345678901234567890"

	result := truncateImageName(input)

	expected := "docker.io/library/nginx:latest12345678901234567..."
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
}

func TestTruncateTagName(t *testing.T) {
	input := "nginx:latest12345678901234567890123456789012345678901234567890"

	result := truncateTagName(input)

	expected := "nginx:latest12345678901234567890123456789012345..."
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
}

func TestPublishOutdatedImages(t *testing.T) {
	mockJS := &sdk.NATSClient{}

	mockresult := model.CheckResultfinal{
		Image:          "test-image",
		Current:        "test-current",
		LatestVersion:  "test-latest",
		VersionsBehind: 23,
		Pod:            "test-pod",
		Namespace:      "test-namespace",
		ClusterName:    "test-cluster",
	}

	mockPublish := gomonkey.ApplyMethod(
		reflect.TypeOf(mockJS),
		"Publish",
		func(*sdk.NATSClient, string, []uint8) error {
			return nil
		},
	)
	defer mockPublish.Reset()

	PublishOutdatedImages(mockresult, mockJS)
}

func TestOutDatedImages(t *testing.T) {
	cases := []struct {
		name              string
		parseImgErr       bool
		listImgErr        bool
		isVersionNegative bool
	}{
		{"success", false, false, false},
		{"parse image error", true, false, false},
		{"list image error", false, true, false},
		{"list image error -1", false, false, true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			fmt.Println("test case", tt.name)

			mockConfig := &rest.Config{}
			mockJS := &sdk.NATSClient{}

			mockInitContainer := "test-initcontainer"
			mockContainer := "test-container"
			mockImages := []model.RunningImage{
				{
					Namespace:     "test-namespace",
					Pod:           "test-pod",
					Image:         "test-image",
					InitContainer: &mockInitContainer,
					Container:     &mockContainer,
					PullableImage: "test-latest",
				},
			}
			patchImages := gomonkey.ApplyFunc(
				ListImages,
				func(config *rest.Config) ([]model.RunningImage, error) {
					if tt.listImgErr {
						return nil, errors.New("list image error")
					}
					return mockImages, nil
				},
			)
			defer patchImages.Reset()

			mockCheckResult1 := model.CheckResult{
				IsAccessible:   true,
				LatestVersion:  "test-latest",
				VersionsBehind: 23,
				CheckError:     "test-error",
				Path:           "test-path",
			}
			mockCheckResult2 := model.CheckResult{
				IsAccessible:   true,
				LatestVersion:  "test-latest",
				VersionsBehind: -1,
				CheckError:     "test-error",
				Path:           "test-path",
			}
			patchcheckresult := gomonkey.ApplyFunc(
				ParseImage,
				func(string, string) (*model.CheckResult, error) {
					if tt.parseImgErr {
						return nil, errors.New("parse image error")
					} else if tt.isVersionNegative {
						return &mockCheckResult2, nil
					}
					return &mockCheckResult1, nil
				},
			)
			defer patchcheckresult.Reset()

			patchImageName := gomonkey.ApplyFunc(
				ParseImageName,
				func(string) (string, string, string, error) {
					if tt.parseImgErr {
						return "", "", "", errors.New("parse image error")
					}
					return "", "", "", nil
				},
			)
			defer patchImageName.Reset()

			mockPublish := gomonkey.ApplyMethod(
				reflect.TypeOf(mockJS),
				"Publish",
				func(*sdk.NATSClient, string, []uint8) error {
					return nil
				},
			)
			defer mockPublish.Reset()

			patchPublishOutdatedImages := gomonkey.ApplyFunc(
				PublishOutdatedImages,
				func(model.CheckResultfinal, *sdk.NATSClient) error {
					return nil
				},
			)
			defer patchPublishOutdatedImages.Reset()

			error := OutDatedImages(mockConfig, mockJS)
			if tt.listImgErr {
				require.Error(t, error)
			} else if tt.parseImgErr {
				require.NoError(t, error)
			} else if tt.isVersionNegative {
				require.NoError(t, error)
			} else {
				require.NoError(t, error)
			}

		})
	}
}

func TestParseImageName(t *testing.T) {
	url, image, tag, err := ParseImageName("redis:4")
	require.NoError(t, err)
	assert.Equal(t, "index.docker.io", url)
	assert.Equal(t, "library/redis", image)
	assert.Equal(t, "4", tag)

	url, image, tag, err = ParseImageName("k8s.gcr.io/cluster-proportional-autoscaler-amd64:1.1.2-r2")
	require.NoError(t, err)
	assert.Equal(t, "k8s.gcr.io", url)
	assert.Equal(t, "library/cluster-proportional-autoscaler-amd64", image)
	assert.Equal(t, "1.1.2-r2", tag)

	url, image, tag, err = ParseImageName("quay.io/coreos/grafana-watcher:v0.0.8")
	require.NoError(t, err)
	assert.Equal(t, "quay.io", url)
	assert.Equal(t, "coreos/grafana-watcher", image)
	assert.Equal(t, "v0.0.8", tag)

	url, image, tag, err = ParseImageName("grafana/grafana:5.0.1")
	require.NoError(t, err)
	assert.Equal(t, "index.docker.io", url)
	assert.Equal(t, "grafana/grafana", image)
	assert.Equal(t, "5.0.1", tag)

	url, image, tag, err = ParseImageName("postgres:10.0")
	require.NoError(t, err)
	assert.Equal(t, "index.docker.io", url)
	assert.Equal(t, "library/postgres", image)
	assert.Equal(t, "10.0", tag)

	url, image, tag, err = ParseImageName("localhost:32000/postgres:10.0")
	require.NoError(t, err)
	assert.Equal(t, "localhost:32000", url)
	assert.Equal(t, "library/postgres", image)
	assert.Equal(t, "10.0", tag)
}

func TestParseImage(t *testing.T) {
	cases := []struct {
		name              string
		parseImgErr       bool
		registryClientErr bool
		tagsErr           bool
		parsTagsErr       bool
	}{
		{"success", false, false, false, false},
		{"parse image error", true, false, false, false},
		{"registry client error", false, true, false, false},
		{"fetch tags error", false, false, true, false},
		{"pars tags error", false, false, false, true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			patchImageName := gomonkey.ApplyFunc(
				ParseImageName,
				func(string) (string, string, string, error) {
					if tt.parseImgErr {
						return "", "", "", errors.New("parse image error")
					}
					return "", "", "", nil
				},
			)
			defer patchImageName.Reset()

			patchRegistryClient := gomonkey.ApplyFunc(
				initRegistryClient,
				func(string) (*registry.Registry, error) {
					if tt.registryClientErr {
						return nil, errors.New("registry client error")
					}
					return &registry.Registry{}, nil
				},
			)
			defer patchRegistryClient.Reset()

			patchTags := gomonkey.ApplyFunc(
				fetchTags,
				func(*registry.Registry, string) ([]string, error) {
					if tt.tagsErr {
						return nil, errors.New("fetch tags error")
					}
					return []string{"test-tag"}, nil
				},
			)
			defer patchTags.Reset()

			mockVersion := semver.Version{}

			patchParseTags := gomonkey.ApplyFunc(
				parseTags,
				func([]string) ([]*semver.Version, []string, error) {
					if tt.parsTagsErr {
						return nil, nil, errors.New("parse tags error")
					}
					return []*semver.Version{
						&mockVersion,
					}, []string{"test-tag"}, nil
				},
			)
			defer patchParseTags.Reset()

			patchNewVersion := gomonkey.ApplyFunc(
				semver.NewVersion,
				func(string) (*semver.Version, error) {
					return &mockVersion, nil
				},
			)
			defer patchNewVersion.Reset()

			patchParseNonServerImage := gomonkey.ApplyFunc(
				parseNonSemverImage,
				func(*registry.Registry, string, string, []string) (*model.CheckResult, error) {
					return &model.CheckResult{}, nil
				},
			)
			defer patchParseNonServerImage.Reset()

			mocksemvertags := []*semver.Version{}
			collection := SemverTagCollection(mocksemvertags)

			mockSemverVersion := &semver.Version{}

			patchRemoveLeastspecific := gomonkey.ApplyMethod(
				reflect.TypeOf(collection),
				"RemoveLeastSpecific",
				func(collection SemverTagCollection) []*semver.Version {
					return []*semver.Version{
						mockSemverVersion,
					}
				},
			)
			defer patchRemoveLeastspecific.Reset()

			patchversionbehind := gomonkey.ApplyMethod(
				reflect.TypeOf(collection),
				"VersionsBehind",
				func(collection SemverTagCollection, detected *semver.Version) ([]*semver.Version, error) {
					return []*semver.Version{}, nil
				},
			)
			defer patchversionbehind.Reset()

			patchresolvetagdata := gomonkey.ApplyFunc(
				resolveTagDates,
				func(*registry.Registry, string, []*semver.Version) ([]*VersionTag, error) {
					return []*VersionTag{}, nil
				},
			)
			defer patchresolvetagdata.Reset()

			resp, err := ParseImage("image", "pullableImage")
			if tt.parseImgErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else if tt.registryClientErr {
				require.NoError(t, err)
			} else if tt.tagsErr {
				require.Error(t, err)
			} else if tt.parsTagsErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, "", resp.LatestVersion)
				require.Equal(t, int64(0), resp.VersionsBehind)
				if resp.LatestVersion == "" {
					require.Equal(t, "", resp.LatestVersion)
					require.Equal(t, int64(0), resp.VersionsBehind)
				} else {
					require.NotEqual(t, "", resp.LatestVersion)
					require.NotEqual(t, int64(-1), resp.VersionsBehind)
				}
			}

		})
	}
}

func TestParseNonSemverImage(t *testing.T) {
	cases := []struct {
		name     string
		dateErr  bool
		parseErr bool
	}{
		{"success", false, false},
		{"date error", true, false},
		{"parse error", false, true},
	}

	for _, tt := range cases {
		mockRegistry := &registry.Registry{}

		patchgetTagDate := gomonkey.ApplyFunc(
			getTagDate,
			func(*registry.Registry, string, string) (string, error) {
				if tt.dateErr {
					return "", errors.New("date error")
				}
				return "", nil
			},
		)
		defer patchgetTagDate.Reset()

		patchParse := gomonkey.ApplyFunc(
			time.Parse,
			func(string, string) (time.Time, error) {
				if tt.parseErr {
					return time.Time{}, errors.New("parse error")
				}
				return time.Now(), nil
			},
		)
		defer patchParse.Reset()

		t.Run(tt.name, func(t *testing.T) {
			resp, err := parseNonSemverImage(mockRegistry, "image", "tag", []string{"test-tag"})
			if tt.dateErr {
				require.NoError(t, err)
			} else if tt.parseErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

func TestInitRegistryClient(t *testing.T) {
	initRegistryClient("docker.io")
}

func TestFetchTags(t *testing.T) {
	mockConfig := &registry.Registry{}
	cases := []struct {
		name    string
		wantErr bool
	}{
		{"success", false},
		{"error", true},
	}

	for _, tt := range cases {
		patchTags := gomonkey.ApplyMethod(
			reflect.TypeOf(mockConfig),
			"Tags",
			func(*registry.Registry, context.Context, string) ([]string, error) {
				if tt.wantErr {
					return nil, errors.New("fetch tags error")
				}
				return []string{"test-tag"}, nil
			},
		)
		defer patchTags.Reset()

		t.Run(tt.name, func(t *testing.T) {
			_, err := fetchTags(mockConfig, "test-image")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseTags(t *testing.T) {
	cases := []struct {
		name    string
		wantErr bool
	}{
		{"success", false},
		{"error", true},
	}

	for _, tt := range cases {

		patchv := gomonkey.ApplyFunc(
			semver.NewVersion,
			func(string) (*semver.Version, error) {
				return nil, errors.New("parse error")
			},
		)
		defer patchv.Reset()

		patchsplit := gomonkey.ApplyFunc(
			splitOutlierSemvers,
			func([]*semver.Version) ([]*semver.Version, []*semver.Version, error) {
				return []*semver.Version{}, []*semver.Version{}, nil
			},
		)

		defer patchsplit.Reset()

		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseTags([]string{"test-tag"})
			if tt.wantErr {
				require.NoError(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSplitOutlierSemvers(t *testing.T) {

	cases := []struct {
		name    string
		wantErr bool
	}{
		{"success", false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := splitOutlierSemvers(makeVersions([]string{"1.0", "1.1", "1.2"}))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSemverTagCollection_Swap(t *testing.T) {
	// Create a sample collection
	collection := SemverTagCollection{
		version.Must(version.NewVersion("1.0.0")),
		version.Must(version.NewVersion("2.0.0")),
		version.Must(version.NewVersion("3.0.0")),
	}

	// Swap elements at indices 0 and 2
	collection.Swap(0, 2)

	// Assert that the elements have been swapped
	assert.Equal(t, "3.0.0", collection[0].String())
	assert.Equal(t, "2.0.0", collection[1].String())
	assert.Equal(t, "1.0.0", collection[2].String())

	// Swap elements at indices 1 and 1 (no change)
	collection.Swap(1, 1)

	// Assert that the collection remains unchanged
	assert.Equal(t, "3.0.0", collection[0].String())
	assert.Equal(t, "2.0.0", collection[1].String())
	assert.Equal(t, "1.0.0", collection[2].String())
}

func makeVersions(versions []string) []*semver.Version {
	allVersions := make([]*semver.Version, 0)
	for _, version := range versions {
		v, _ := semver.NewVersion(version)
		allVersions = append(allVersions, v)
	}
	return allVersions
}

func makeOriginal(versions []*semver.Version) []string {
	allVersions := make([]string, 0)
	for _, version := range versions {
		v := version.Original()
		allVersions = append(allVersions, v)
	}
	return allVersions
}

func TestRemoveLeastSpecific(t *testing.T) {
	tests := []struct {
		name           string
		versions       []string
		expectVersions []string
	}{
		{
			name:           "major",
			versions:       []string{"1.0", "1.1", "1.2", "1"},
			expectVersions: []string{"1.0", "1.1", "1.2"},
		},
		{
			name:           "zeros",
			versions:       []string{"1.0", "1"},
			expectVersions: []string{"1.0"},
		},
		{
			name:           "minor",
			versions:       []string{"0.1.1", "0.1.2", "0.1"},
			expectVersions: []string{"0.1.1", "0.1.2"},
		},
		{
			name:           "patch",
			versions:       []string{"0.2.1", "0.2.2", "0.2.3"},
			expectVersions: []string{"0.2.1", "0.2.2", "0.2.3"},
		},
		{
			name:           "similar version numbers",
			versions:       []string{"0.11.1", "0.11", "11.0"},
			expectVersions: []string{"0.11.1", "11.0"},
		},
		{
			name:           "different versions",
			versions:       []string{"0.1", "2.1"},
			expectVersions: []string{"0.1", "2.1"},
		},
		{
			name:           "include last",
			versions:       []string{"0.1", "0.2.1", "0.3.4", "0.4"},
			expectVersions: []string{"0.1", "0.2.1", "0.3.4", "0.4"},
		},
		{
			name:           "variety",
			versions:       []string{"0.1.0", "0.1", "0.2.0", "0.2", "0.10.0", "0.10", "0.11.0", "0.11", "0.13.0", "0.13.1", "0.13.2", "0.13.3", "0.13", "0.17.0", "0.17.1", "0.17", "0.18.0", "0.18", "0.21.0", "0.21", "0"},
			expectVersions: []string{"0.1.0", "0.2.0", "0.10.0", "0.11.0", "0.13.0", "0.13.1", "0.13.2", "0.13.3", "0.17.0", "0.17.1", "0.18.0", "0.21.0"},
		},
		{
			name:           "preserve major version",
			versions:       []string{"0.1", "0.2.1", "1", "2", "3.5", "3", "4"},
			expectVersions: []string{"0.1", "0.2.1", "1", "2", "3.5", "4"},
		},
		{
			name:           "variations",
			versions:       []string{"1.0.0", "1.0", "1"},
			expectVersions: []string{"1.0.0"},
		},
		{
			name:           "variations 2",
			versions:       []string{"0.0.0", "0.0", "0"},
			expectVersions: []string{"0.0.0"},
		},
		{
			name:           "more segments",
			versions:       []string{"3.5.1.1", "3.5.1", "4.5.1"},
			expectVersions: []string{"3.5.1.1", "4.5.1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			semverTags := makeVersions(test.versions)
			collection := SemverTagCollection(semverTags)
			tags := collection.RemoveLeastSpecific()

			originalRemoved := makeOriginal(tags)
			require.EqualValues(t, test.expectVersions, originalRemoved)
		})
	}
}

func TestTagCollectionUnique(t *testing.T) {
	tests := []struct {
		name           string
		versions       []string
		expectVersions []string
	}{
		{
			name:           "tagged versions",
			versions:       []string{"1.0.1", "1.0.2", "1.0.1-alpine", "1.0.1-debian"},
			expectVersions: []string{"1.0.1", "1.0.2"},
		},
		{
			name:           "tagged major version",
			versions:       []string{"4-alpine", "4"},
			expectVersions: []string{"4"},
		},
		{
			name:           "different major and minor versions",
			versions:       []string{"10.4", "9"},
			expectVersions: []string{"9", "10.4"},
		},
		{
			name:           "different major only versions",
			versions:       []string{"10", "11"},
			expectVersions: []string{"10", "11"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vers := makeVersions(test.versions)
			uniqueVers, err := SemverTagCollection(vers).Unique()
			require.NoError(t, err)
			actualVers := makeOriginal(uniqueVers)

			require.Equal(t, test.expectVersions, actualVers)
		})
	}
}

func TestTagCollectionSort(t *testing.T) {
	tests := []struct {
		name           string
		versions       []string
		expectVersions []string
	}{
		{
			name:           "same major versions",
			versions:       []string{"10", "10.4"},
			expectVersions: []string{"10", "10.4"},
		},
		{
			name:           "different major versions",
			versions:       []string{"10", "11.1"},
			expectVersions: []string{"10", "11.1"},
		},
		{
			name:           "same major and minor versions",
			versions:       []string{"9.1.3", "9.1.0", "9.1.4", "9.1"},
			expectVersions: []string{"9.1.0", "9.1", "9.1.3", "9.1.4"},
		},
		{
			name:           "different major and minor versions",
			versions:       []string{"10.1.2", "10.0", "10", "10.3.2", "11", "10.1.3", "10.1"},
			expectVersions: []string{"10.0", "10", "10.1", "10.1.2", "10.1.3", "10.3.2", "11"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vers := makeVersions(test.versions)
			sort.Sort(SemverTagCollection(vers))
			actualVers := makeOriginal(vers)

			require.Equal(t, test.expectVersions, actualVers)
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		expect   int
	}{
		{
			name:     "major versions match",
			versions: []string{"10", "10"},
			expect:   0,
		},
		{
			name:     "major version is less than",
			versions: []string{"9", "10"},
			expect:   -1,
		},
		{
			name:     "minor versions is less than",
			versions: []string{"10.1", "10"},
			expect:   1,
		},
		{
			name:     "minor versions is greater than",
			versions: []string{"10.3", "10.1"},
			expect:   1,
		},
		{
			name:     "minor versions match",
			versions: []string{"10.1", "10.1"},
			expect:   0,
		},
		{
			name:     "patch versions is less than",
			versions: []string{"10.1.2", "10.1.3"},
			expect:   -1,
		},
		{
			name:     "patch versions is greater than",
			versions: []string{"10.1.4", "10.1.3"},
			expect:   1,
		},
		{
			name:     "patch versions match",
			versions: []string{"10.1.2", "10.1.2"},
			expect:   0,
		},
		{
			name:     "major version only with patch",
			versions: []string{"10", "10.1.2"},
			expect:   -1,
		},
		{
			name:     "minor version greater, patch version less",
			versions: []string{"10.2.3", "10.1.4"},
			expect:   1,
		},
		{
			name:     "major version less, minor version greater",
			versions: []string{"9.1.2", "10.0.1"},
			expect:   -1,
		},
		{
			name:     "major version less, also shorter",
			versions: []string{"9.1", "10.2.2"},
			expect:   -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vers := makeVersions(test.versions)

			actual := compareVersions(vers[0], vers[1])
			require.Equal(t, test.expect, actual)

			reverse := compareVersions(vers[1], vers[0])
			require.Equal(t, -test.expect, reverse)
		})
	}
}

func TestResolveTagDates(t *testing.T) {
	tests := []struct {
		name      string
		hostname  string
		imageName string
		versions  []string
	}{
		{
			name:      "postgres",
			hostname:  "index.docker.io",
			imageName: "library/postgres",
			versions:  []string{"10.0", "10.1", "10.2"},
		},
		{
			name:      "tiller",
			hostname:  "gcr.io",
			imageName: "kubernetes-helm/tiller",
			versions:  []string{"v2.14.1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			allVersions := makeVersions(test.versions)
			mockreg := &registry.Registry{}

			patchgetTagDate := gomonkey.ApplyFunc(
				getTagDate,
				func(reg *registry.Registry, imageName string, versionFromTag string) (string, error) {
					return "2006-01-02T15:04:05Z", nil
				},
			)
			defer patchgetTagDate.Reset()

			versionTags, err := resolveTagDates(mockreg, test.imageName, allVersions)
			req.NoError(err)

			for _, versionTag := range versionTags {
				_, err = time.Parse(time.RFC3339, versionTag.Date)
				req.NoError(err)
			}
		})
	}
}

func TestGetTagDate(t *testing.T) {
	cases := []struct {
		name         string
		wantErr      bool
		unmarshalErr bool
	}{
		{"success", false, false},
		{"error", true, false},
		{"unmarshal error 2", false, true},
	}

	for _, tt := range cases {

		mockRegistry := &registry.Registry{
			Username: "test-username",
			Password: "test-password",
			Domain:   "test-domain",
			URL:      "test-url",
			Client:   &http.Client{},
			Logf:     func(string, ...interface{}) {},
		}

		mockManifest := schema1.SignedManifest{}
		mockManifest.History = []schema1.History{
			{V1Compatibility: "history-1"},
		}

		patchmanifest := gomonkey.ApplyMethod(
			reflect.TypeOf(mockRegistry),
			"ManifestV1",
			func(*registry.Registry, context.Context, string, string) (schema1.SignedManifest, error) {
				if tt.wantErr {
					return schema1.SignedManifest{}, errors.New("manifest error")
				}
				return mockManifest, nil
			},
		)
		defer patchmanifest.Reset()

		patchUnmarshal := gomonkey.ApplyFunc(
			json.Unmarshal,
			func([]byte, interface{}) error {
				if tt.unmarshalErr {
					return errors.New("unmarshal error")
				}
				return nil
			},
		)
		defer patchUnmarshal.Reset()

		t.Run(tt.name, func(t *testing.T) {
			resp, err := getTagDate(mockRegistry, "test-image", "test-tag")
			if tt.wantErr {
				require.Error(t, err)
			} else if tt.unmarshalErr {
				require.Error(t, err)
			} else {
				if resp == "" {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}

func TestVersionsBehind(t *testing.T) {
	v1, _ := semver.NewVersion("1.0.0")
	v2, _ := semver.NewVersion("1.1.0")
	v3, _ := semver.NewVersion("1.2.3")

	collection := SemverTagCollection{v1, v2, v3}

	currentVersion := v1
	expectedResult := []*semver.Version{v1, v2, v3}
	result, err := collection.VersionsBehind(currentVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)

	currentVersion = v2
	expectedResult = []*semver.Version{v2, v3}
	result, err = collection.VersionsBehind(currentVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)

	currentVersion = v3
	expectedResult = []*semver.Version{v3}
	result, err = collection.VersionsBehind(currentVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)

	currentVersion, _ = semver.NewVersion("2.0.0")
	expectedResult = []*semver.Version{currentVersion}
	result, err = collection.VersionsBehind(currentVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestListImages(t *testing.T) {
	cases := []struct {
		name         string
		namespaceErr bool
	}{
		{"namespace error", true},
	}

	for _, tt := range cases {
		mockConfig := &rest.Config{}

		t.Run(tt.name, func(t *testing.T) {
			_, err := ListImages(mockConfig)
			fmt.Println(err)
			if tt.namespaceErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})

	}
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
	return nil, nil
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
