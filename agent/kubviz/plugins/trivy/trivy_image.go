package trivy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	exec "os/exec"
	"strings"

	"github.com/aquasecurity/trivy/pkg/types"
	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type JetStreamContextInterface interface {
	nats.JetStream
	nats.JetStreamManager
	nats.KeyValueManager
	nats.ObjectStoreManager
	AccountInfo(opts ...nats.JSOpt) (*nats.AccountInfo, error)
}

func RunTrivyImageScans(config *rest.Config, js nats.JetStreamContext) error {
	pvcMountPath := "/mnt/agent/kbz"
	trivyImageCacheDir := fmt.Sprintf("%s/trivy-imagecache", pvcMountPath)
	err := os.MkdirAll(trivyImageCacheDir, 0755)
	if err != nil {
		log.Printf("Error creating Trivy Image cache directory: %v\n", err)
		return err
	}
	// clearCacheCmd := "trivy image --clear-cache"

	ctx := context.Background()
	tracer := otel.Tracer("trivy-image")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "RunTrivyImageScans")
	span.SetAttributes(attribute.String("trivy-image-scan-agent", "image-scan"))
	defer span.End()

	images, err := ListImages(config)
	if err != nil {
		log.Println("error occured while trying to list images, error :", err.Error())
		return err
	}

	for _, image := range images {
		var report types.Report
		scanCmd := fmt.Sprintf("trivy image %s --timeout 60m -f json -q --cache-dir %s", image.PullableImage, trivyImageCacheDir)
		out, err := executeTrivyImage(scanCmd)
		if err != nil {
			log.Printf("Error scanning image %s: %v", image.PullableImage, err)
			continue // Move on to the next image in case of an error
		}

		parts := strings.SplitN(string(out), "{", 2)
		if len(parts) <= 1 {
			log.Println("No output from image scan command", err)
			continue // Move on to the next image if there's no output
		}

		// log.Println("Command logs for image", parts[0])
		jsonPart := "{" + parts[1]
		// log.Println("First 200 image scan lines output", jsonPart[:200])
		// log.Println("Last 200 image scan lines output", jsonPart[len(jsonPart)-200:])

		err = json.Unmarshal([]byte(jsonPart), &report)
		if err != nil {
			log.Printf("Error occurred while Unmarshalling json for image: %v", err)
			continue // Move on to the next image in case of an error
		}
		// _, err = executeCommandTrivy(clearCacheCmd)
		// if err != nil {
		// 	log.Printf("Error executing command: %v\n", err)
		// 	return err
		// }
		err = PublishImageScanReports(report, js)
		if err != nil {
			return err
		}
	}
	return nil
}

func PublishImageScanReports(report types.Report, js nats.JetStreamContext) error {
	metrics := model.TrivyImage{
		ID:          uuid.New().String(),
		ClusterName: ClusterName,
		Report:      report,
	}
	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(constants.TRIVY_IMAGE_SUBJECT, metricsJson)
	if err != nil {
		return err
	}
	log.Printf("Trivy image report with ID:%s has been published\n", metrics.ID)
	return nil
}
func executeTrivyImage(command string) ([]byte, error) {

	// ctx := context.Background()
	// tracer := otel.Tracer("trivy-image")
	// _, span := tracer.Start(opentelemetry.BuildContext(ctx), "executeCommandTrivyImage")
	// span.SetAttributes(attribute.String("trivy-image-agent", "trivyimage-command-running"))
	// defer span.End()

	cmd := exec.Command("/bin/sh", "-c", command)
	var outc, errc bytes.Buffer
	cmd.Stdout = &outc
	cmd.Stderr = &errc
	err := cmd.Run()
	// if outc.Len() > 0 {
	// 	log.Printf("Command Output: %s\n", outc.String())
	// }
	if errc.Len() > 0 {
		log.Printf("Command Error: %s\n", errc.String())
	}
	if err != nil {
		return nil, fmt.Errorf("error while executing trivy image command: %v", err)
	}
	return outc.Bytes(), err
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
