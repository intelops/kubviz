package application

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/intelops/kubviz/model"
	v1 "github.com/vijeyash1/go-github-container/v1"
)

func (app *Application) GithubContainerWatch() {
	if app.GithubConfig.Org == "" {
		log.Println("Aborting the github container monitoring process , please provide a github organisation under env GITHUB_ORG")
		return
	}
	if app.GithubConfig.Token == "" {
		log.Println("Aborting the github container monitoring process , please provide a github token  under env GITHUB_TOKEN")
		return
	}
	client := v1.NewGithubClient(app.GithubConfig.Org, app.GithubConfig.Token)
	packages, err := client.FetchPackages()
	if err != nil {
		log.Printf("Error fetching packages: %v\n", err)
		return
	}
	for _, pkg := range packages {
		versions, err := client.FetchVersions(pkg.Name)
		if err != nil {
			log.Printf("Error fetching versions for package %s: %v\n", pkg.Name, err)
			continue
		}
		for _, version := range versions {
			image := BuildImageDetails(pkg, version)
			data, err := json.Marshal(image)
			if err != nil {
				log.Printf("unable to marshal the image details %v", err)
				return
			}
			err = app.conn.Publish(data, "github")
			if err != nil {
				log.Printf("Publish failed for event: %v, reason: %v", string(data), err)
			}
		}
	}
}

func BuildImageDetails(pkg v1.Package, version v1.Version) model.GithubImage {
	return model.GithubImage{
		PackageId:    fmt.Sprint(pkg.ID),
		CreatedAt:    version.CreatedAt,
		ImageName:    pkg.Name,
		Organisation: pkg.Owner.Login,
		UpdatedAt:    version.UpdatedAt,
		Visibility:   pkg.Visibility,
		ShaID:        version.Name,
		ImageId:      fmt.Sprint(version.ID),
	}
}
