package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("/upload-file", func(c *gin.Context) {
		file, _ := c.FormFile("file")
		c.SaveUploadedFile(file, file.Filename)
		c.JSON(http.StatusOK, gin.H{
			"message": "File uploaded successfully",
		})

	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
