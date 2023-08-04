package handler

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// parse errors
var (
	ErrReadingBody   = errors.New("error reading the request body")
	ErrPublishToNats = errors.New("error while publishing to nats")
)

func (ah *APIHandler) PostEventDockerHub(c *gin.Context) {
	defer func() {
		_, _ = io.Copy(io.Discard, c.Request.Body)
		_ = c.Request.Body.Close()
	}()
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil || len(payload) == 0 {
		log.Printf("%v: %v", ErrReadingBody, err)
		c.Status(http.StatusBadRequest)
		return
	}
	log.Printf("Received event from docker artifactory: %v", string(payload))
	err = ah.conn.Publish(payload, "Dockerhub_Registry")
	if err != nil {
		log.Printf("%v: %v", ErrPublishToNats, err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}
