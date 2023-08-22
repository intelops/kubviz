package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
	"k8s.io/client-go/rest"
)

func publishTrivySbomReport(report model.Sbom, js nats.JetStreamContext, errCh chan error) {
	metrics := model.Reports{
		ID:     uuid.New().String(),
		Report: report,
	}
	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(constants.TRIVY_SBOM_SUBJECT, metricsJson)
	if err != nil {
		errCh <- err
	}

	log.Printf("Trivy report with BomFormat:%v has been published\n", metrics.Report.BomFormat)
	errCh <- nil
}

func executeCommandSbom(ctx context.Context, command string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println("Execute Command Error", err.Error())
	}

	return stdout, nil
}

// func RunTrivySbomScan(config *rest.Config, js nats.JetStreamContext, wg *sync.WaitGroup, errCh chan error) {
// 	defer wg.Done()
// 	images, err := ListImages(config)
// 	log.Println("length of images", len(images))

// 	if err != nil {
// 		log.Printf("failed to list images: %v", err)
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 700*time.Second)
// 	defer cancel()

// 	var wgc sync.WaitGroup
// 	wgc.Add(len(images)) // Set the wait group count to the number of images

// 	for i, image := range images {
// 		fmt.Printf("pullable Image %#v\n", image.PullableImage)

// 		// Start a goroutine for each image
// 		go func(i int, image model.RunningImage) {
// 			defer wgc.Done()

// 			// Execute the Trivy command with the context
// 			command := fmt.Sprintf("trivy image --format cyclonedx %s", image.PullableImage)
// 			out, err := executeCommandSbom(ctx, command)

// 			if ctx.Err() == context.DeadlineExceeded {
// 				log.Printf("Command execution timeout for image %s", image.PullableImage)
// 				return // Move on to the next image
// 			}

// 			if err != nil {
// 				log.Printf("Error executing Trivy for image %s: %v", image.PullableImage, err)
// 				return // Move on to the next image in case of an error
// 			}

// 			// Check if the output is empty or invalid JSON
// 			if len(out) == 0 {
// 				log.Printf("Trivy output is empty for image %s", image.PullableImage)
// 				return // Move on to the next image
// 			}

// 			// Extract the JSON data from the output
// 			var report model.Sbom
// 			err = json.Unmarshal(out, &report)
// 			if err != nil {
// 				log.Printf("Error unmarshaling JSON data for image %s: %v", image.PullableImage, err)
// 				return // Move on to the next image in case of an error
// 			}

// 			// Publish the report using the given function
// 			publishTrivySbomReport(report, js, errCh)
// 		}(i, image)
// 	}

//		// Wait for all the goroutines to complete
//		wgc.Wait()
//	}
func RunTrivySbomScan(config *rest.Config, js nats.JetStreamContext, wg *sync.WaitGroup, errCh chan error) {
	log.Println("trivy run started****************")
	defer wg.Done()
	images, err := ListImages(config)
	log.Println("length of images", len(images))

	if err != nil {
		log.Printf("failed to list images: %v", err)
	}

	ctx := context.Background()

	for _, image := range images {
		fmt.Printf("pullable Image %#v\n", image.PullableImage)

		// Execute the Trivy command with the context
		command := fmt.Sprintf("trivy -d image --format cyclonedx %s", image.PullableImage)
		out, err := executeCommandSbom(ctx, command)

		if err != nil {
			log.Printf("Error executing Trivy for image %s: %v", image.PullableImage, err)
			continue // Move on to the next image in case of an error
		}

		// Check if the output is empty or invalid JSON
		if len(out) == 0 {
			log.Printf("Trivy output is empty for image %s", image.PullableImage)
			continue // Move on to the next image
		}

		// Extract the JSON data from the output
		var report model.Sbom
		err = json.Unmarshal(out, &report)
		if err != nil {
			log.Printf("Error unmarshaling JSON data for image %s: %v", image.PullableImage, err)
			continue // Move on to the next image in case of an error
		}

		// Publish the report using the given function
		publishTrivySbomReport(report, js, errCh)
	}
}
