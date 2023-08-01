package application

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/intelops/kubviz/model"
	v1 "github.com/vijeyash1/go-github-container/v1"
)

// GithubContainerWatch monitors and publishes container image details from a specified GitHub organization's repositories.
func (app *Application) GithubContainerWatch() {
	// Check if the GitHub organization is provided in the configuration.
	if app.GithubConfig.Org == "" {
		log.Println("Aborting the GitHub container monitoring process, please provide a GitHub organization under env GITHUB_ORG")
		return
	}

	// Check if the GitHub token is provided in the configuration.
	if app.GithubConfig.Token == "" {
		log.Println("Aborting the GitHub container monitoring process, please provide a GitHub token under env GITHUB_TOKEN")
		return
	}

	// Create a new GitHub client with the provided organization and token.
	client := v1.NewGithubClient(app.GithubConfig.Org, app.GithubConfig.Token)

	// Fetch the list of packages (repositories) from the GitHub organization.
	packages, err := client.FetchPackages()
	if err != nil {
		log.Printf("Error fetching packages: %v\n", err)
		return
	}

	// Iterate through each package and its versions to build and publish container image details.
	for _, pkg := range packages {
		// Fetch the list of versions for the current package.
		versions, err := client.FetchVersions(pkg.Name)
		if err != nil {
			log.Printf("Error fetching versions for package %s: %v\n", pkg.Name, err)
			continue // Skip this package and proceed with the next one.
		}

		// Iterate through each version of the package and build container image details.
		for _, version := range versions {
			image := BuildImageDetails(pkg, version) // Construct container image details from package and version.
			data, err := json.Marshal(image)         // Serialize the image details to JSON.
			if err != nil {
				log.Printf("Unable to marshal the image details: %v", err)
				return // Abort the monitoring process in case of serialization error.
			}

			// Publish the JSON-encoded image details to the "github" topic.
			err = app.conn.Publish(data, "Github_Registry")
			if err != nil {
				log.Printf("Publish failed for event: %v, reason: %v", string(data), err)
			}
		}
	}
}

// BuildImageDetails constructs a model.GithubImage from the given v1.Package and v1.Version.
func BuildImageDetails(pkg v1.Package, version v1.Version) model.GithubImage {
	// Create and return a new GithubImage object using the provided package and version information.
	return model.GithubImage{
		PackageId:    fmt.Sprint(pkg.ID),     // Convert the package ID to a string and set it as PackageId.
		CreatedAt:    version.CreatedAt,      // Set the creation timestamp of the version as CreatedAt.
		ImageName:    pkg.Name,               // Set the package name as ImageName.
		Organisation: pkg.Owner.Login,        // Set the GitHub organization or owner login as Organisation.
		UpdatedAt:    version.UpdatedAt,      // Set the last update timestamp of the version as UpdatedAt.
		Visibility:   pkg.Visibility,         // Set the visibility (public, private, etc.) of the package as Visibility.
		ShaID:        version.Name,           // Set the version name as ShaID (can be a commit SHA or version tag).
		ImageId:      fmt.Sprint(version.ID), // Convert the version ID to a string and set it as ImageId.
	}
}
