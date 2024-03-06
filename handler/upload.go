package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type UploadController struct{}

type Request struct {
	FileName   string `json:"file_name"`
	ChunkCount int    `json:"chunk_count"`
	ChunkSize  int    `json:"chunk_size"`
}

func (u UploadController) RequestUpload(c *gin.Context) {
	var request Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Request received",
	})
}

func (u UploadController) UploadChunk(c *gin.Context) {
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
	// create file if not exists
	var file *os.File
	_, err = os.Stat("uploaded_file")
	if os.IsNotExist(err) {
		file, err = os.Create("uploaded_file")
		if err != nil {
			c.JSON(500, gin.H{"error": "Error creating file"})
			return
		}
	} else {
		file, err = os.OpenFile("uploaded_file", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error opening file"})
			return
		}
	}
	defer file.Close()
	chunk := c.Request.Body

	writer := io.MultiWriter(file)
	_, err = io.CopyN(writer, chunk, end-start+1)
	if err != nil {
		if err != io.EOF {
			c.JSON(500, gin.H{"error": "Error writing chunk to file"})
			return
		}
	}

	fmt.Println("Chunk " + strconv.FormatInt(end, 10) + " uploaded successfully")

	if end == total-1 {
		c.JSON(http.StatusOK, gin.H{
			"message": "File uploaded successfully",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Chunk " + strconv.FormatInt(end, 10) + " uploaded successfully",
		})
	}
}

func (u UploadController) CompleteUpload(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
	})
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
