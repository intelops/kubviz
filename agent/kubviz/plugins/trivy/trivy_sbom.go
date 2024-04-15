package trivy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PublishTrivySbomReport(report map[string]interface{}, js nats.JetStreamContext) error {

	metrics := model.Sbom{
		ID:          uuid.New().String(),
		ClusterName: ClusterName,
		Report:      report,
	}
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		log.Println("error occurred while marshalling sbom metrics in agent", err.Error())
		return err
	}
	_, err = js.Publish(constants.TRIVY_SBOM_SUBJECT, metricsJSON)
	if err != nil {
		return err
	}
	log.Printf("Trivy sbom report with Id %v has been published\n", metrics.ID)
	return nil
}

func executeCommandSbom(command string) ([]byte, error) {

	cmd := exec.Command("/bin/sh", "-c", command)
	var outc, errc bytes.Buffer
	cmd.Stdout = &outc
	cmd.Stderr = &errc
	err := cmd.Run()
	if err != nil {
		log.Println("Execute SBOM Command Error", err.Error())
	}
	return outc.Bytes(), err
}

func RunTrivySbomScan(config *rest.Config, js nats.JetStreamContext) error {
	log.Println("trivy sbom scan started...")
	pvcMountPath := "/mnt/agent/kbz"
	trivySbomCacheDir := fmt.Sprintf("%s/trivy-sbomcache", pvcMountPath)
	err := os.MkdirAll(trivySbomCacheDir, 0755)
	if err != nil {
		log.Printf("Error creating Trivy cache directory: %v\n", err)
		return err
	}

	ctx := context.Background()
	tracer := otel.Tracer("trivy-sbom")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "RunTrivySbomScan")
	defer span.End()

	images, err := ListImagesforSbom(config)

	if err != nil {
		log.Printf("failed to list images: %v", err)
	}
	for _, image := range images {

		sbomcmd := fmt.Sprintf("trivy image --format cyclonedx %s --cache-dir %s", image.PullableImage, trivySbomCacheDir)
		out, err := executeCommandSbom(sbomcmd)

		if err != nil {
			log.Printf("Error executing Trivy for image sbom %s: %v", image.PullableImage, err)
			continue // Move on to the next image in case of an error
		}
		if out == nil {
			log.Printf("Trivy output is nil for image sbom %s", image.PullableImage)
			continue
		}
		// Check if the output is empty or invalid JSON
		if len(out) == 0 {
			log.Printf("Trivy output is empty for image sbom %s", image.PullableImage)
			continue // Move on to the next image
		}

		var report map[string]interface{}
		err = json.Unmarshal(out, &report)
		if err != nil {
			log.Printf("Error unmarshaling JSON data for image sbom %s: %v", image.PullableImage, err)
			continue // Move on to the next image in case of an error
		}
		err = PublishTrivySbomReport(report, js)
		if err != nil {
			log.Printf("Error publishing Trivy SBOM report for image %s: %v", image.PullableImage, err)
			continue
		}
	}
	return nil
}

func ListImagesforSbom(config *rest.Config) ([]model.RunningImage, error) {
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
				pullable = strings.TrimPrefix(pullable, "docker-pullable://")
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
				pullable = strings.TrimPrefix(pullable, "docker-pullable://")

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
