package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	engine := gin.Default()
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	v1engine := engine.Group("/api/v1")

	v1engine.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	v1engine.POST("/upload-file", func(c *gin.Context) {

		// Header Content-Range: bytes 0-999999/100000000
		headerContentRange := c.Request.Header.Get("Content-Range")
		if headerContentRange == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Content-Range header is required",
			})
			return
		}

		// Parse the Content-Range header to extract the start and end byte positions and the total file size
		start, end, total, err := parseContentRange(headerContentRange)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid Content-Range header"})
			return
		}
		file, err := os.Create("uploaded_file")
		if err != nil {
			c.JSON(500, gin.H{"error": "Error creating file"})
			return
		}
		defer file.Close()
		writer := io.MultiWriter(file)
		chunk := c.Request.Body
		_, err = io.CopyN(writer, chunk, end-start+1)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error writing chunk to file"})
			return
		}

		if end == total-1 {
			c.JSON(http.StatusOK, gin.H{
				"message": "File uploaded successfully",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Chunk " + strconv.FormatInt(end, 10) + " uploaded successfully",
			})
		}

	})
	engine.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func parseContentRange(header string) (int64, int64, int64, error) {
	// Parse the Content-Range header to extract the start and end byte positions and the total file size
	// Example header: "bytes 0-999999/100000000"
	// Return the start, end, and total byte positions as int64 values, and an error if the header is invalid

	// Split the header string into its components
	parts := strings.Split(header, " ")
	if len(parts) != 3 || parts[0] != "bytes" {
		return 0, 0, 0, fmt.Errorf("invalid Content-Range header format: %s", header)
	}

	// Parse the start and end byte positions and the total file size from the header
	rangeParts := strings.Split(parts[1], "-")
	start, err := strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid start byte position: %s", rangeParts[0])
	}
	end, err := strconv.ParseInt(rangeParts[1], 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid end byte position: %s", rangeParts[1])
	}
	total, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid total file size: %s", parts[2])
	}

	return start, end, total, nil
}
