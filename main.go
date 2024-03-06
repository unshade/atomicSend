package main

import (
	"atomicSend/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func main() {
	engine := gin.Default()
	/*client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})*/

	v1engine := engine.Group("/api/v1")
	config := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Content-Range", "Upload-Id"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
		AllowOrigins:     []string{"*"},
	}
	v1engine.Use(cors.New(config))
	engine.Use(cors.New(config))

	v1engine.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	upload := handler.UploadController{}
	v1engine.POST("/request-upload", upload.RequestUpload)
	v1engine.POST("/upload-chunk", upload.UploadChunk)
	v1engine.POST("/finalize-upload", upload.CompleteUpload)

	engine.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
