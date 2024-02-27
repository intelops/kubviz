package trivy

import (
	"bytes"
	// "context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	exec "os/exec"
	"strings"

	"github.com/aquasecurity/trivy/pkg/types"
	"github.com/google/uuid"
	"github.com/intelops/kubviz/agent/kubviz/plugins/outdated"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	// "github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/nats-io/nats.go"
	// "go.opentelemetry.io/otel"
	// "go.opentelemetry.io/otel/attribute"
	"k8s.io/client-go/rest"
)

func RunTrivyImageScans(config *rest.Config, js nats.JetStreamContext) error {
	pvcMountPath := "/mnt/agent/kbz"
	trivyImageCacheDir := fmt.Sprintf("%s/trivy-imagecache", pvcMountPath)
	err := os.MkdirAll(trivyImageCacheDir, 0777)
	if err != nil {
		log.Printf("Error creating Trivy Image cache directory: %v\n", err)
		return err
	}
	// clearCacheCmd := "trivy image --clear-cache"

	// ctx := context.Background()
	// tracer := otel.Tracer("trivy-image")
	// _, span := tracer.Start(opentelemetry.BuildContext(ctx), "RunTrivyImageScans")
	// span.SetAttributes(attribute.String("trivy-image-scan-agent", "image-scan"))
	// defer span.End()

	images, err := outdated.ListImages(config)
	if err != nil {
		log.Println("error occured while trying to list images, error :", err.Error())
		return err
	}

	for i, image := range images {
		var report types.Report
		scanCmd := fmt.Sprintf("trivy image %s --timeout 60m -f json -q --cache-dir %s", image.PullableImage, trivyImageCacheDir)
		out, err := executeTrivyImage(scanCmd)
		if err != nil {
			log.Printf("Error scanning image with index %v : %s: %v", i, image.PullableImage, err)
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
		return nil, err
	}
	return outc.Bytes(), err
}
