package kubepreupgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/davecgh/go-spew/spew"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var ClusterName string = os.Getenv("CLUSTER_NAME")

const (
	baseURL         = "https://raw.githubusercontent.com/kubernetes/kubernetes"
	fileURL         = "api/openapi-spec/swagger.json"
	crdGroup        = "apiextensions.k8s.io"
	apiRegistration = "apiregistration.k8s.io"
)

type ignoreStruct map[string]struct{}

type groupResourceKind struct {
	GroupVersion string
	ResourceName string
	ResourceKind string
}

type ResourceStruct struct {
	GroupVersion, ResourceName string
}

type PreferredResource map[string]ResourceStruct

var (
	k8sVersion             = "master"
	deletedApiReplacements = map[string]groupResourceKind{
		"extensions/v1beta1/Ingress": {"networking.k8s.io/v1", "ingresses", "Ingress"},
	}
)
var result *model.Result

func publishK8sDepricated_Deleted_Api(result *model.Result, js nats.JetStreamContext) error {
	for _, deprecatedAPI := range result.DeprecatedAPIs {
		deprecatedAPI.ClusterName = ClusterName
		deprecatedAPIJson, _ := json.Marshal(deprecatedAPI)
		_, err := js.Publish(constants.EventSubject_depricated, deprecatedAPIJson)
		if err != nil {
			return err
		}
	}

	for _, deletedAPI := range result.DeletedAPIs {
		deletedAPI.ClusterName = ClusterName
		fmt.Println("deletedAPI", deletedAPI)
		deletedAPIJson, _ := json.Marshal(deletedAPI)
		_, err := js.Publish(constants.EventSubject_deleted, deletedAPIJson)
		if err != nil {
			return err
		}
	}

	log.Printf("Metrics with Deletedapi and depricated api has been published")
	return nil
}

func KubePreUpgradeDetector(config *rest.Config, js nats.JetStreamContext) error {

	ctx := context.Background()
	tracer := otel.Tracer("kubepreupgrade")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "KubePreUpgradeDetector")
	span.SetAttributes(attribute.String("kubepug-plugin-agent", "kubepug-output"))
	defer span.End()

	pvcMountPath := "/mnt/agent/kbz"
	uniqueDir := fmt.Sprintf("%s/kubepug", pvcMountPath)
	err := os.MkdirAll(uniqueDir, 0755)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("%s/swagger-%s.json", uniqueDir, k8sVersion)
	url := fmt.Sprintf("%s/%s/%s", baseURL, k8sVersion, fileURL)
	err = downloadFile(filename, url)
	if err != nil {
		return err
	}
	defer os.RemoveAll(filename)
	kubernetesAPIs, err := PopulateKubeAPIMap(filename)
	if err != nil {
		return err
	}
	result = getResults(config, kubernetesAPIs)
	err = publishK8sDepricated_Deleted_Api(result, js)
	return err
}

func PopulateKubeAPIMap(swagfile string) (model.KubernetesAPIs, error) {
	var kubeAPIs = make(model.KubernetesAPIs)
	// log.Infof("Populating the PopulateKubeAPIMap")
	jsonFile, err := os.Open(swagfile)
	if err != nil {
		log.Error(err)
	}
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	err = jsonFile.Close()
	if err != nil {
		return nil, err
	}
	var definitionsMap map[string]interface{}
	err = json.Unmarshal(byteValue, &definitionsMap)
	if err != nil {
		return nil, err
	}
	definitions := definitionsMap["definitions"].(map[string]interface{})
	for k, value := range definitions {
		val := value.(map[string]interface{})
		if kubeapivalue, valid := getKubeAPIValues(val); valid {
			log.Debugf("Valid API object found for %s", k)

			var name string
			if kubeapivalue.Group != "" {
				name = fmt.Sprintf("%s/%s/%s", kubeapivalue.Group, kubeapivalue.Version, kubeapivalue.Kind)
			} else {
				name = fmt.Sprintf("%s/%s", kubeapivalue.Version, kubeapivalue.Kind)
			}

			log.Debugf("Adding %s to map. Deprecated: %t", name, kubeapivalue.Deprecated)
			kubeAPIs[name] = kubeapivalue
		}
	}
	return kubeAPIs, nil
}
func getKubeAPIValues(value map[string]interface{}) (model.KubeAPI, bool) {
	var (
		valid, deprecated                 bool
		description, group, version, kind string
	)

	gvk, valid, err := unstructured.NestedSlice(value, "x-kubernetes-group-version-kind")
	if !valid || err != nil {
		return model.KubeAPI{}, false
	}
	gvkMap := gvk[0]
	group, version, kind = getGroupVersionKind(gvkMap.(map[string]interface{}))

	description, found, err := unstructured.NestedString(value, "description")

	if !found || err != nil || description == "" {
		log.Debugf("Marking the resource as invalid because it doesn't contain a description")
		return model.KubeAPI{}, false
	}

	if strings.Contains(strings.ToLower(description), "deprecated") {
		log.Debugf("API Definition contains the word DEPRECATED in its description")
		deprecated = true
	}

	if valid {
		return model.KubeAPI{
			Description: description,
			Group:       group,
			Kind:        kind,
			Version:     version,
			Deprecated:  deprecated,
		}, true
	}

	return model.KubeAPI{}, false
}
func downloadFile(filename, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Error(err)
		return err
	}
	if resp.StatusCode > 305 {
		log.Errorf("could not download the swagger file %s", url)
		return fmt.Errorf("failed to download file, status code: %d", resp.StatusCode)
	}
	contentLength := resp.ContentLength
	log.Infof("The size of the file to be downloaded for kubepreupgrade plugin is %d bytes", contentLength)

	defer resp.Body.Close()
	out, err := os.Create(filename)
	if err != nil {
		log.Error(err)
	}
	defer out.Close()
	bytesCopied, err := io.Copy(out, resp.Body)
	if err != nil {
		log.WithError(err).Error("Failed to copy the file contents")
		return err
	}
	log.Infof("Downloaded %d bytes for file %s", bytesCopied, filename)

	return nil
}

func getGroupVersionKind(value map[string]interface{}) (group, version, kind string) {
	for k, v := range value {
		switch k {
		case "group":
			group = v.(string) //nolint: errcheck
		case "version":
			version = v.(string) //nolint: errcheck
		case "kind":
			kind = v.(string) //nolint: errcheck
		}
	}

	return group, version, kind
}

func getResults(configRest *rest.Config, kubeAPIs model.KubernetesAPIs) *model.Result {
	var res model.Result
	var deleted []model.DeletedAPI
	var deprecated []model.DeprecatedAPI
	var resourceName string

	client, err := dynamic.NewForConfig(configRest)
	if err != nil {
		log.Errorf("Failed to create the K8s client while listing Deprecated objects: %s", err)
	}

	disco, err := discovery.NewDiscoveryClientForConfig(configRest)
	if err != nil {
		log.Errorf("Failed to create the K8s Discovery client: %s", err)
	}

	ResourceAndGV := DiscoverResourceNameAndPreferredGV(disco)
	for _, dpa := range kubeAPIs {
		// We only want deprecated APIs :)
		if !dpa.Deprecated {
			continue
		}

		group, version, kind := dpa.Group, dpa.Version, dpa.Kind
		var gvk string

		if group != "" {
			gvk = fmt.Sprintf("%s/%s/%s", group, version, kind)
		} else {
			gvk = fmt.Sprintf("%s/%s", version, kind)
		}

		if _, ok := ResourceAndGV[gvk]; !ok {
			log.Debugf("Skipping the resource %s because it doesn't exists in the APIServer", gvk)
			continue
		}

		prefResource := ResourceAndGV[gvk]

		if prefResource.ResourceName == "" || prefResource.GroupVersion == "" {
			log.Debugf("Skipping the resource %s because it doesn't exists in the APIServer", gvk)
			continue
		}

		gv, err := schema.ParseGroupVersion(prefResource.GroupVersion)
		if err != nil {
			log.Warnf("Failed to parse GroupVersion %s of resource %s existing in the API Server: %s", prefResource.GroupVersion, prefResource.ResourceName, err)
		}

		gvrPreferred := gv.WithResource(prefResource.ResourceName)

		log.Debugf("Listing objects for %s/%s/%s", group, version, prefResource.ResourceName)
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: prefResource.ResourceName}
		list, err := client.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
		if apierrors.IsNotFound(err) {
			continue
		}

		if apierrors.IsForbidden(err) {
			log.Errorf("Failed to list objects in the cluster. Permission denied! Please check if you have the proper authorization")
		}

		if err != nil {
			log.Errorf("Failed communicating with k8s while listing objects. \nError: %v", err)
		}

		// Now let's see if there's a preferred API containing the same objects
		if gvr != gvrPreferred {
			log.Infof("Listing objects for Preferred %s/%s", prefResource.GroupVersion, prefResource.ResourceName)

			listPref, err := client.Resource(gvrPreferred).List(context.TODO(), metav1.ListOptions{})
			if apierrors.IsForbidden(err) {
				log.Errorf("Failed to list objects in the cluster. Permission denied! Please check if you have the proper authorization")
			}

			if err != nil && !apierrors.IsNotFound(err) {
				log.Errorf("Failed communicating with k8s while listing objects. \nError: %v", err)
			}
			// If len of the lists is the same we can "assume" they're the same list
			if len(list.Items) == len(listPref.Items) {
				log.Infof("%s/%s/%s contains the same length of %d items that preferred %s/%s with %d items, skipping", group, version, kind, len(list.Items), prefResource.GroupVersion, kind, len(listPref.Items))
				continue
			}
		}

		if len(list.Items) > 0 {
			log.Infof("Found %d deprecated objects of type %s/%s/%s", len(list.Items), group, version, resourceName)
			api := model.DeprecatedAPI{
				Kind:        kind,
				Deprecated:  dpa.Deprecated,
				Group:       group,
				Name:        resourceName,
				Version:     version,
				Description: dpa.Description,
			}

			api.Items = ListObjects(list.Items)
			deprecated = append(deprecated, api)
		}
	}
	res.DeprecatedAPIs = deprecated
	resourcesList, err := disco.ServerPreferredResources()
	if err != nil {
		if apierrors.IsForbidden(err) {
			log.Errorf("Failed to list Server Resources. Permission denied! Please check if you have the proper authorization")
		}

		log.Errorf("Failed communicating with k8s while discovering server resources. \nError: %v", err)
	}
	var ignoreObjects ignoreStruct = make(map[string]struct{})
	for _, resources := range resourcesList {
		if strings.Contains(resources.GroupVersion, crdGroup) {
			version := strings.Split(resources.GroupVersion, "/")[1]
			populateCRDGroups(client, version, ignoreObjects)
		}

		if strings.Contains(resources.GroupVersion, apiRegistration) {
			version := strings.Split(resources.GroupVersion, "/")[1]
			populateAPIService(client, version, ignoreObjects)
		}
	}
	for _, resourceGroupVersion := range resourcesList {
		// We don't want CRDs or APIExtensions to be walked
		if _, ok := ignoreObjects[strings.Split(resourceGroupVersion.GroupVersion, "/")[0]]; ok {
			continue
		}

		for i := range resourceGroupVersion.APIResources {
			resource := &resourceGroupVersion.APIResources[i] // We don't want to check subObjects (like pods/status)
			if len(strings.Split(resource.Name, "/")) != 1 {
				continue
			}

			keyAPI := fmt.Sprintf("%s/%s", resourceGroupVersion.GroupVersion, resource.Kind)
			if _, ok := kubeAPIs[keyAPI]; !ok {

				gvr, list := getResources(client, groupResourceKind{resourceGroupVersion.GroupVersion, resource.Name, resource.Kind})

				if newApi, ok := deletedApiReplacements[keyAPI]; ok {
					list.Items = fixDeletedItemsList(client, list.Items, newApi)
				}

				if len(list.Items) > 0 {
					log.Debugf("Found %d deleted items in %s/%s", len(list.Items), gvr.Group, resource.Kind)
					d := model.DeletedAPI{
						Deleted: true,
						Name:    resource.Name,
						Group:   gvr.Group,
						Kind:    resource.Kind,
						Version: gvr.Version,
					}

					d.Items = ListObjects(list.Items)
					deleted = append(deleted, d)
				}
			}
		}
	}
	res.DeletedAPIs = deleted
	return &res
}
func populateCRDGroups(dynClient dynamic.Interface, version string, ignoreStruct ignoreStruct) {
	crdgvr := schema.GroupVersionResource{
		Group:    crdGroup,
		Version:  version,
		Resource: "customresourcedefinitions",
	}
	crdList, err := dynClient.Resource(crdgvr).List(context.TODO(), metav1.ListOptions{})
	if apierrors.IsNotFound(err) {
		return
	}
	if err != nil {
		log.Errorf("Failed to connect to K8s cluster to List CRDs: %s", err)
	}
	var empty struct{}
	for _, d := range crdList.Items {
		group, found, err := unstructured.NestedString(d.Object, "spec", "group")
		// No group fields found, move on!
		if err != nil || !found {
			continue
		}
		if _, ok := ignoreStruct[group]; !ok {
			ignoreStruct[group] = empty
		}
	}
}
func populateAPIService(dynClient dynamic.Interface, version string, ignoreStruct ignoreStruct) {
	apisvcgvr := schema.GroupVersionResource{
		Group:    apiRegistration,
		Version:  version,
		Resource: "apiservices",
	}
	apisvcList, err := dynClient.Resource(apisvcgvr).List(context.TODO(), metav1.ListOptions{})
	if apierrors.IsNotFound(err) {
		return
	}
	if err != nil {
		log.Errorf("Failed to connect to K8s cluster to List API Services: %s", err)
	}
	var empty struct{}
	for _, d := range apisvcList.Items {
		_, foundSvc, errSvc := unstructured.NestedString(d.Object, "spec", "service", "name")
		group, foundGrp, errGrp := unstructured.NestedString(d.Object, "spec", "group")
		// No services fields or group field found, move on!
		if errSvc != nil || !foundSvc || errGrp != nil || !foundGrp {
			continue
		}

		if _, ok := ignoreStruct[group]; !ok {
			ignoreStruct[group] = empty
		}
	}
}

func DiscoverResourceNameAndPreferredGV(client *discovery.DiscoveryClient) PreferredResource {
	pr := make(PreferredResource)

	resourcelist, err := client.ServerPreferredResources()
	if err != nil {
		if apierrors.IsNotFound(err) {
			return pr
		}
		if apierrors.IsForbidden(err) {
			log.Errorf("Failed to list objects for Name discovery. Permission denied! Please check if you have the proper authorization")
		}

		log.Errorf("Failed communicating with k8s while discovering the object preferred name and gv. Error: %v", err)
	}

	for _, rl := range resourcelist {
		for i := range rl.APIResources {
			item := ResourceStruct{
				GroupVersion: rl.GroupVersion,
				ResourceName: rl.APIResources[i].Name,
			}

			gvk := fmt.Sprintf("%v/%v", rl.GroupVersion, rl.APIResources[i].Kind)
			pr[gvk] = item
		}
	}

	return pr
}
func ListObjects(items []unstructured.Unstructured) (deprecatedItems []model.Item) {
	for _, d := range items {
		name := d.GetName()
		namespace := d.GetNamespace()
		if namespace != "" {
			deprecatedItems = append(deprecatedItems, model.Item{Scope: "OBJECT", ObjectName: name, Namespace: namespace})
		} else {
			deprecatedItems = append(deprecatedItems, model.Item{Scope: "GLOBAL", ObjectName: name})
		}
	}

	return deprecatedItems
}
func getResources(dynClient dynamic.Interface, grk groupResourceKind) (schema.GroupVersionResource, *unstructured.UnstructuredList) {

	gv, err := schema.ParseGroupVersion(grk.GroupVersion)
	if err != nil {
		log.Errorf("Failed to Parse GroupVersion of Resource: %s", err)
	}

	gvr := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: grk.ResourceName}
	list, err := dynClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if apierrors.IsNotFound(err) || apierrors.IsMethodNotSupported(err) {
		return gvr, list
	}

	if apierrors.IsForbidden(err) {
		log.Errorf("Failed to list Server Resources of type %s/%s/%s. Permission denied! Please check if you have the proper authorization", gv.Group, gv.Version, grk.ResourceKind)
	}

	if err != nil {
		log.Errorf("Failed to List objects of type %s/%s/%s. \nError: %v", gv.Group, gv.Version, grk.ResourceKind, err)
	}

	return gvr, list
}

func fixDeletedItemsList(dynClient dynamic.Interface, oldApiItems []unstructured.Unstructured, grk groupResourceKind) []unstructured.Unstructured {

	_, newApiItems := getResources(dynClient, grk)
	newApiItemsMap := make(map[string]bool)

	for _, item := range newApiItems.Items {
		uid := spew.Sprint(item.Object["metadata"].(map[string]interface{})["uid"])
		newApiItemsMap[uid] = true
	}

	deletedItems := []unstructured.Unstructured{}
	for _, item := range oldApiItems {
		uid := spew.Sprint(item.Object["metadata"].(map[string]interface{})["uid"])
		// Only adds to the deleted list if not found in the new API list
		if !newApiItemsMap[uid] {
			deletedItems = append(deletedItems, item)
		}
	}
	return deletedItems
}
