package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"blog/service"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
)

const (
	PORT = 1234
)

func initRouter() *gin.Engine {
	router := gin.New()
	router.Delims("{{", "}}")
	router.Use(gin.Recovery())
	router.Use(service.ParseIndex())
	router.LoadHTMLGlob("./template/*.html")
	router.GET("/", service.Home)
	router.GET("/about", service.About)

	router.GET("/post/:name", service.GetPost)
	router.GET("/category/:name", service.GetCategory)
	router.GET("/tag/:name")
	return router
}

func main() {
	router := initRouter()
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(PORT),
		WriteTimeout: 20 * time.Second,
		Handler:      router,
	}
	err := gracehttp.Serve(server)
	if err != nil {
		log.Fatal("服务启动失败:", err.Error())
	}
}
