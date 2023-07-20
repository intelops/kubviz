package application

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intelops/kubviz/agent/git/api"
)

func (app *Application) PostGitea(c *gin.Context) {
	repo := "Gitea"
	event := c.Request.Header.Get("X-Gitea-Event")
	if len(event) == 0 {
		log.Println("error getting the gitea event from header")
		return
	}
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
	}
	app.conn.Publish(jsonData, repo, event)
}

func (app *Application) PostAzure(c *gin.Context) {

	repo := "Azure"
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
	}
	app.conn.Publish(jsonData, repo, "azure event")
}

// githubHandler handles the github webhooks post requests.
func (app *Application) PostGithub(c *gin.Context) {
	repo := "Github"
	event := c.Request.Header.Get("X-GitHub-Event")
	if len(event) == 0 {
		log.Println("error getting the github event from header")
		return
	}
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
	}
	app.conn.Publish(jsonData, repo, event)
}

// gitlabHandler handles the github webhooks post requests.
func (app *Application) PostGitlab(c *gin.Context) {

	repo := "Gitlab"
	event := c.Request.Header.Get("X-Gitlab-Event")
	if len(event) == 0 {
		log.Println("error getting the gitlab event from header")
		return
	}
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
	}
	app.conn.Publish(jsonData, repo, event)
}

// bitBucketHandler handles the github webhooks post requests.
func (app *Application) PostBitbucket(c *gin.Context) {

	repo := "BitBucket"
	event := c.Request.Header.Get("X-Event-Key")
	if len(event) == 0 {
		log.Println("error getting the bitbucket event from header")
		return
	}
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
	}
	app.conn.Publish(jsonData, repo, event)
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
