package application

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intelops/kubviz/agent/git/api"
	"github.com/intelops/kubviz/gitmodels/azuremodel"
	"github.com/intelops/kubviz/model"
)

func (app *Application) PostGitea(c *gin.Context) {

	event := c.Request.Header.Get(string(model.GiteaHeader))
	if len(event) == 0 {
		log.Println("error getting the gitea event from header")
		return
	}
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
	}
	app.conn.Publish(jsonData, string(model.GiteaProvider), model.GiteaHeader, model.EventValue(event))
}

func (app *Application) PostAzure(c *gin.Context) {

	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
		c.Status(http.StatusInternalServerError)
		return
	}
	var pl azuremodel.BasicEvent
	err = json.Unmarshal([]byte(jsonData), &pl)
	if err != nil {
		log.Println("Error Reading Request Body")
		c.Status(http.StatusInternalServerError)
		return
	}
	event := pl.EventType
	if string(event) == "" {
		log.Println("Error Reading Request Body")
		c.Status(http.StatusInternalServerError)
		return
	}
	app.conn.Publish(jsonData, string(model.AzureDevopsProvider), model.AzureHeader, model.EventValue(event))
}

// githubHandler handles the github webhooks post requests.
func (app *Application) PostGithub(c *gin.Context) {
	event := c.Request.Header.Get(string(model.GithubHeader))
	if len(event) == 0 {
		log.Println("error getting the github event from header")
		c.Status(http.StatusBadRequest)
		return
	}
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
		c.Status(http.StatusInternalServerError)
		return
	}
	app.conn.Publish(jsonData, string(model.GithubProvider), model.GithubHeader, model.EventValue(event))
}

// gitlabHandler handles the github webhooks post requests.
func (app *Application) PostGitlab(c *gin.Context) {

	event := c.Request.Header.Get(string(model.GitlabHeader))
	if len(event) == 0 {
		log.Println("error getting the gitlab event from header")
		c.Status(http.StatusBadRequest)
		return
	}
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
		c.Status(http.StatusInternalServerError)
		return
	}
	app.conn.Publish(jsonData, string(model.GitlabProvider), model.GitlabHeader, model.EventValue(event))
}

// bitBucketHandler handles the github webhooks post requests.
func (app *Application) PostBitbucket(c *gin.Context) {

	event := c.Request.Header.Get(string(model.BitBucketHeader))
	if len(event) == 0 {
		log.Println("error getting the bitbucket event from header")
		c.Status(http.StatusBadRequest)
		return
	}
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("Error Reading Request Body")
		c.Status(http.StatusInternalServerError)
		return
	}
	app.conn.Publish(jsonData, string(model.BitBucketProvider), model.BitBucketHeader, model.EventValue(event))
}

// GetLiveness handles the liveness check get requests.
func (app *Application) GetLiveness(c *gin.Context) {
	c.Status(http.StatusOK)
}

// GetApiDocs handles the GetApiDocs get requests.
func (app *Application) GetApiDocs(c *gin.Context) {
	swagger, err := api.GetSwagger()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, swagger)
}
