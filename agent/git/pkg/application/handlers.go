package application

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kube-tarian/kubviz/agent/git/api"
)

func (app *Application) PostGitea(c *gin.Context) {
	repo := "Gitea"
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
	}
	app.conn.Publish(jsonData, repo)
}

func (app *Application) PostAzure(c *gin.Context) {

	repo := "Azure"
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
	}
	app.conn.Publish(jsonData, repo)
}

// githubHandler handles the github webhooks post requests.
func (app *Application) PostGithub(c *gin.Context) {
	repo := "Github"
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
	}
	app.conn.Publish(jsonData, repo)
}

// gitlabHandler handles the github webhooks post requests.
func (app *Application) PostGitlab(c *gin.Context) {
	repo := "Gitlab"
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
	}
	app.conn.Publish(jsonData, repo)
}

// bitBucketHandler handles the github webhooks post requests.
func (app *Application) PostBitbucket(c *gin.Context) {
	repo := "BitBucket"
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
	}
	app.conn.Publish(jsonData, repo)
}
func (app *Application) GetLiveness(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (app *Application) GetApiDocs(c *gin.Context) {
	swagger, err := api.GetSwagger()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, swagger)
}
