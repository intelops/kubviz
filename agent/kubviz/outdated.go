package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/intelops/kubviz/constants"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"

	"github.com/docker/docker/api/types"
	"github.com/genuinetools/reg/registry"
	semver "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	maxImageLength = 50
	maxTagLength   = 50
)

var (
	dockerImageNameRegex = regexp.MustCompile("(?:([^\\/]+)\\/)?(?:([^\\/]+)\\/)?([^@:\\/]+)(?:[@:](.+))")
)
var (
	hashedusername = os.Getenv("DOCKER_USERNAME")
	hashedpassword = os.Getenv("DOCKER_PASSWORD")
)

func truncateImageName(imageName string) string {
	truncatedImageName := imageName
	if len(truncatedImageName) > maxImageLength {
		truncatedImageName = fmt.Sprintf("%s...", truncatedImageName[0:maxImageLength-3])
	}
	return truncatedImageName
}
func truncateTagName(tagName string) string {
	truncatedTagName := tagName
	if len(tagName) > maxTagLength {
		truncatedTagName = fmt.Sprintf("%s...", truncatedTagName[0:maxTagLength-3])
	}
	return truncatedTagName
}
func PublishOutdatedImages(out model.CheckResultfinal, js nats.JetStreamContext) error {
	metrics := out
	metrics.ClusterName = ClusterName
	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(constants.EventSubject_outdated_images, metricsJson)
	if err != nil {
		return err
	}
	log.Printf("Metrics with outdated images has been published")
	return nil
}

func outDatedImages(config *rest.Config, js nats.JetStreamContext) error {
	images, err := ListImages(config)
	if err != nil {
		log.Println("unable to list images")
		return err
	}
	for _, image := range images {
		namespace := image.Namespace
		pod := image.Pod
		checkResult, _ := ParseImage(image.Image, image.PullableImage)
		repo, img, tag, err := ParseImageName(image.Image)
		final := model.CheckResultfinal{}
		if err != nil {
			imageName := fmt.Sprintf("%s/%s", repo, img)
			img := truncateImageName(imageName)
			final.Image = img
			message := "Unable to get image data"
			if checkResult != nil {
				message = checkResult.CheckError
			}
			final.LatestVersion = message
			final.Namespace = namespace
			final.Pod = pod
			err := PublishOutdatedImages(final, js)
			if err != nil {
				return err
			}
		} else {
			if checkResult != nil {
				if checkResult.VersionsBehind != -1 {
					tagtrunk := truncateTagName(tag)
					final.Current = tagtrunk
					imageName := fmt.Sprintf("%s/%s", repo, img)
					img := truncateImageName(imageName)
					final.Image = img
					final.LatestVersion = checkResult.LatestVersion
					final.VersionsBehind = checkResult.VersionsBehind
					final.Namespace = namespace
					final.Pod = pod
					err := PublishOutdatedImages(final, js)
					if err != nil {
						return err
					}
				} else {
					tagtrunk := truncateTagName(tag)
					final.Current = tagtrunk
					imageName := fmt.Sprintf("%s/%s", repo, img)
					img := truncateImageName(imageName)
					final.Image = img
					message := "Unable to get image data"
					if checkResult != nil {
						message = checkResult.CheckError
					}
					final.LatestVersion = message
					final.Namespace = namespace
					final.Pod = pod
					err := PublishOutdatedImages(final, js)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func ParseImageName(imageName string) (string, string, string, error) {
	matches := dockerImageNameRegex.FindStringSubmatch(imageName)

	if len(matches) != 5 {
		return "", "", "", fmt.Errorf("Expected 5 matches in regex, but found %d", len(matches))
	}

	hostname := matches[1]
	namespace := matches[2]
	image := matches[3]
	tag := matches[4]

	if namespace == "" && hostname != "" {
		if !strings.Contains(hostname, ".") && !strings.Contains(hostname, ":") {
			namespace = hostname
			hostname = ""
		}
	}

	if hostname == "" {
		hostname = "index.docker.io"
	}

	if namespace == "" {
		namespace = "library"
	}

	return hostname, fmt.Sprintf("%s/%s", namespace, image), tag, nil
}
func ListImages(config *rest.Config) ([]model.RunningImage, error) {
	var err error
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create clientset")
	}
	ctx := context.Background()
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list namespaces")
	}

	runningImages := []model.RunningImage{}
	for _, namespace := range namespaces.Items {
		pods, err := clientset.CoreV1().Pods(namespace.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list pods")
		}

		for _, pod := range pods.Items {
			for _, initContainerStatus := range pod.Status.InitContainerStatuses {
				pullable := initContainerStatus.ImageID
				if strings.HasPrefix(pullable, "docker-pullable://") {
					pullable = strings.TrimPrefix(pullable, "docker-pullable://")
				}
				runningImage := model.RunningImage{
					Pod:           pod.Name,
					Namespace:     pod.Namespace,
					InitContainer: &initContainerStatus.Name,
					Image:         initContainerStatus.Image,
					PullableImage: pullable,
				}
				runningImages = append(runningImages, runningImage)
			}

			for _, containerStatus := range pod.Status.ContainerStatuses {
				pullable := containerStatus.ImageID
				if strings.HasPrefix(pullable, "docker-pullable://") {
					pullable = strings.TrimPrefix(pullable, "docker-pullable://")
				}
				runningImage := model.RunningImage{
					Pod:           pod.Name,
					Namespace:     pod.Namespace,
					Container:     &containerStatus.Name,
					Image:         containerStatus.Image,
					PullableImage: pullable,
				}
				runningImages = append(runningImages, runningImage)
			}
		}
	}

	// Remove exact duplicates
	cleanedImages := []model.RunningImage{}
	seenImages := make(map[string]bool)
	for _, runningImage := range runningImages {
		if !seenImages[runningImage.PullableImage] {
			cleanedImages = append(cleanedImages, runningImage)
			seenImages[runningImage.PullableImage] = true
		}
	}

	return cleanedImages, nil
}

func ParseImage(image string, pullableImage string) (*model.CheckResult, error) {
	hostname, imageName, tag, err := ParseImageName(image)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse image name")
	}

	reg, err := initRegistryClient(hostname)
	if err != nil {
		return &model.CheckResult{
			IsAccessible:   false,
			LatestVersion:  "",
			VersionsBehind: -1,
			CheckError:     fmt.Sprintf("Cannot access registry: %s", err.Error()),
		}, nil
	}

	tags, err := fetchTags(reg, imageName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch image tags")
	}

	semverTags, nonSemverTags, err := parseTags(tags)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse tags")
	}

	detectedSemver, err := semver.NewVersion(tag)
	if err != nil {
		return parseNonSemverImage(reg, imageName, tag, nonSemverTags)
	}

	// From here on, we can assume that we are on a semver tag
	semverTags = append(semverTags, detectedSemver)
	collection := SemverTagCollection(semverTags)

	versionsBehind, err := collection.VersionsBehind(detectedSemver)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate versions behind")
	}
	trueVersionsBehind := SemverTagCollection(versionsBehind).RemoveLeastSpecific()

	behind := len(trueVersionsBehind) - 1

	checkResult := model.CheckResult{
		IsAccessible: true,
	}
	checkResult.VersionsBehind = int64(behind)

	versionPaths, err := resolveTagDates(reg, imageName, trueVersionsBehind)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve tag dates")
	}
	path, err := json.Marshal(versionPaths)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal version path")
	}
	checkResult.Path = string(path)

	checkResult.LatestVersion = trueVersionsBehind[len(trueVersionsBehind)-1].String()

	return &checkResult, nil
}

func parseNonSemverImage(reg *registry.Registry, imageName string, tag string, nonSemverTags []string) (*model.CheckResult, error) {
	laterDates := []string{}
	tagDate, err := getTagDate(reg, imageName, tag)
	if err != nil {
		return &model.CheckResult{
			IsAccessible:   true,
			LatestVersion:  tag,
			VersionsBehind: -1,
			CheckError:     "Unable to determine date from current tag",
		}, nil
	}
	myDate, err := time.Parse(time.RFC3339Nano, tagDate)
	if err != nil {
		return nil, err
	}

	for _, nonSemverTag := range nonSemverTags {
		otherDate, err := getTagDate(reg, imageName, nonSemverTag)
		if err != nil {
			continue
		}

		o, err := time.Parse(time.RFC3339Nano, otherDate)
		if err != nil {
			continue
		}
		if o.After(myDate) {
			laterDates = append(laterDates, otherDate)
		}
	}

	behind := int64(len(laterDates))
	return &model.CheckResult{
		IsAccessible:   true,
		LatestVersion:  tag,
		VersionsBehind: behind,
		CheckError:     "",
	}, nil
}

const (
	// SemverOutlierMajorVersionThreshold defines the number of major versions that must be skipped before
	// the next version is considered an outlier
	// setting this to 2 allows only 1 major version to be skipped
	SemverOutlierMajorVersionThreshold = 2
)

func initRegistryClient(hostname string) (*registry.Registry, error) {
	if hostname == "docker.io" {
		hostname = "index.docker.io"
	}
	var useAuth bool
	var (
		decodedusername, decodedpassword []byte
		err                              error
	)
	var auth types.AuthConfig
	username := ""
	password := ""
	if hostname == "index.docker.io" {
		useAuth = false
		decodedusername, err = base64.StdEncoding.DecodeString(hashedusername)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode username from env")
		}
		decodedpassword, err = base64.StdEncoding.DecodeString(hashedpassword)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode password from env")
		}
		useAuth = true
	}
	if useAuth {
		username = string(decodedusername)
		password = string(decodedpassword)
	}

	auth = types.AuthConfig{
		Username:      username,
		Password:      password,
		ServerAddress: hostname,
	}

	reg, err := registry.New(context.TODO(), auth, registry.Opt{
		SkipPing: true,
		Timeout:  time.Duration(time.Second * 5),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create registry client")
	}

	return reg, nil
}

func fetchTags(reg *registry.Registry, imageName string) ([]string, error) {
	tags, err := reg.Tags(context.TODO(), imageName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list tags")
	}

	return tags, nil
}

func parseTags(tags []string) ([]*semver.Version, []string, error) {
	semverTags := make([]*semver.Version, 0, 0)
	nonSemverTags := make([]string, 0, 0)

	for _, tag := range tags {
		v, err := semver.NewVersion(tag)
		if err != nil {
			nonSemverTags = append(nonSemverTags, tag)
		} else {
			semverTags = append(semverTags, v)
		}
	}

	// some semver tags might be outliers and should be treated as non-semver tags
	// For more info, see https://github.com/replicatedhq/outdated/issues/19
	outlierSemver, remainingSemver, err := splitOutlierSemvers(semverTags)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to split outliers")
	}

	for _, outlier := range outlierSemver {
		nonSemverTags = append(nonSemverTags, outlier.String())
	}

	return remainingSemver, nonSemverTags, nil
}

func splitOutlierSemvers(allSemverTags []*semver.Version) ([]*semver.Version, []*semver.Version, error) {
	if len(allSemverTags) == 0 {
		return []*semver.Version{}, []*semver.Version{}, nil
	}

	sortable := SemverTagCollection(allSemverTags)
	sort.Sort(sortable)

	outliers := []*semver.Version{}
	remaining := []*semver.Version{}

	lastVersion := allSemverTags[0]
	isInOutlier := false
	for _, v := range allSemverTags {
		if v.Segments()[0]-lastVersion.Segments()[0] > SemverOutlierMajorVersionThreshold {
			isInOutlier = true
		}

		if isInOutlier {
			outliers = append(outliers, v)
		} else {
			remaining = append(remaining, v)
		}

		lastVersion = v
	}

	return outliers, remaining, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}

type VersionTag struct {
	Sort    int    `json:"sort"`
	Version string `json:"version"`
	Date    string `json:"date"`
}

type V1History struct {
	Created string `json:"created,omitempty"`
}

type SemverTagCollection []*semver.Version

func (c SemverTagCollection) Len() int {
	return len(c)
}

func (c SemverTagCollection) Less(i, j int) bool {
	return compareVersions(c[i], c[j]) < 0
}

func compareVersions(verI *semver.Version, verJ *semver.Version) int {
	if verI.LessThan(verJ) {
		return -1
	} else if verI.GreaterThan(verJ) {
		return 1
	}

	return 0
}

func (c SemverTagCollection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c SemverTagCollection) VersionsBehind(currentVersion *semver.Version) ([]*semver.Version, error) {
	cleaned, err := c.Unique()
	if err != nil {
		return []*semver.Version{}, errors.Wrap(err, "failed to deduplicate versions")
	}

	sortable := SemverTagCollection(cleaned)
	sort.Sort(sortable)

	for idx := range sortable {
		if sortable[idx].Original() == currentVersion.Original() {
			return sortable[idx:], nil
		}
	}

	return []*semver.Version{
		currentVersion,
	}, nil // /shrug
}

// Unique will create a new sorted slice with the same versions that have different tags removed.
// While this is valid in semver, it's used in docker images differently
// For example: redis:4-alpine and redis:4-debian are the same version
func (c SemverTagCollection) Unique() ([]*semver.Version, error) {
	unique := make(map[string]*semver.Version)

	for _, v := range c {
		var ver string
		var validSegments []int
		splitTag := strings.Split(v.Original(), ".")
		segments := v.Segments()

		if len(splitTag) == 1 {
			validSegments = []int{segments[0]}
		} else if len(splitTag) == 2 {
			validSegments = segments[0:2]
		} else if len(splitTag) == 3 {
			validSegments = segments
		}

		strSegments := []string{}
		for _, segment := range validSegments {
			strSegments = append(strSegments, strconv.Itoa(segment))
		}
		ver = strings.Join(strSegments, ".")

		if _, exists := unique[ver]; !exists {
			unique[ver] = v
		} else {
			// we want the shortest tag -
			// e.g. between redis:4-alpine and redis:4, we want redis:4
			if len(v.Original()) < len(unique[ver].Original()) {
				unique[ver] = v
			}
		}
	}

	result := make([]*semver.Version, 0, 0)
	for _, u := range unique {
		result = append(result, u)
	}

	sort.Sort(SemverTagCollection(result))

	return result, nil
}

// RemoveLeastSpecific given a sorted collection will remove the least specific version
func (c SemverTagCollection) RemoveLeastSpecific() []*semver.Version {
	if c.Len() == 0 {
		return []*semver.Version{}
	}

	cleanedVersions := []*semver.Version{c[0]}
	for i := 0; i < len(c)-1; i++ {
		j := i + 1
		iSegments := c[i].Segments()
		jSegments := c[j].Segments()

		isLessSpecific := true
		for idx, iSegment := range iSegments {
			if len(jSegments) < idx+1 {
				break
			}
			if iSegment > 0 && jSegments[idx] == 0 {
				break
			}
			if iSegment != jSegments[idx] {
				isLessSpecific = false
				break
			}
		}

		if !isLessSpecific {
			cleanedVersions = append(cleanedVersions, c[j])
		}
	}

	return cleanedVersions
}

func resolveTagDates(reg *registry.Registry, imageName string, sortedVersions []*semver.Version) ([]*VersionTag, error) {
	var wg sync.WaitGroup
	var mux sync.Mutex
	versionTags := make([]*VersionTag, 0)

	wg.Add(len(sortedVersions))
	for idx, version := range sortedVersions {
		versionFromTag := version.Original()
		versionTag := VersionTag{
			Sort:    idx,
			Version: versionFromTag,
		}

		go func(versionFromTag string) {
			date, err := getTagDate(reg, imageName, versionFromTag)
			if err == nil {
				versionTag.Date = date
			}

			mux.Lock()
			versionTags = append(versionTags, &versionTag)
			mux.Unlock()

			wg.Done()
		}(versionFromTag)

	}
	wg.Wait()

	return versionTags, nil
}

func getTagDate(reg *registry.Registry, imageName string, versionFromTag string) (string, error) {
	manifest, err := reg.ManifestV1(context.TODO(), imageName, versionFromTag)
	if err != nil {
		return "", errors.Wrap(err, "unable to get manifest from image")
	}
	for _, history := range manifest.History {
		v1History := V1History{}
		err := json.Unmarshal([]byte(history.V1Compatibility), &v1History)
		if err != nil {
			// if it doesn't fit...throw it away
			continue
		}
		if v1History.Created != "" {
			return v1History.Created, nil
		}
	}

	return "", errors.New("no dates found")
}
