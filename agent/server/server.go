package server

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
)

func EnableProfile(r *gin.Engine) {
	pprofGroup := r.Group("/debug/pprof")
	{
		pprofGroup.GET("/", gin.WrapH(http.DefaultServeMux))
		pprofGroup.GET("/cmdline", gin.WrapH(http.DefaultServeMux))
		pprofGroup.GET("/profile", gin.WrapH(http.DefaultServeMux))
		pprofGroup.POST("/symbol", gin.WrapH(http.DefaultServeMux))
		pprofGroup.GET("/symbol", gin.WrapH(http.DefaultServeMux))
		pprofGroup.GET("/trace", gin.WrapH(http.DefaultServeMux))
	}
	r.GET("/liveness", func(c *gin.Context) {
		c.String(http.StatusOK, "Alive")
	})
}

func StartServer() {
	r := gin.Default()
	EnableProfile(r)
	log.Fatal(r.Run(":8080"))
}
