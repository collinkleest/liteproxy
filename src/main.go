package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func addCorsHeaders(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	c.Header("Access-Control-Allow-Credentials", "true")
}

func decompressBody(body io.ReadCloser, encoding string) (io.ReadCloser, error) {
	if encoding == "gzip" {
		gzipReader, err := gzip.NewReader(body)
		if err != nil {
			return nil, err
		}
		return gzipReader, nil
	}
	return body, nil
}

func handlePreflightRequest(c *gin.Context) {
	addCorsHeaders(c)
	c.Status(http.StatusOK)
}

func handleProxyRequest(c *gin.Context) {
	target := c.Query("url")
	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target query param is missing"})
		return
	}

	req, err := http.NewRequest(c.Request.Method, target, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header = c.Request.Header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to forward request"})
		return
	}

	defer resp.Body.Close()

	encoding := resp.Header.Get("Content-Encoding")
	resp.Body, err = decompressBody(resp.Body, encoding)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decompress response"})
		return
	}

	addCorsHeaders(c)
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

func handleRequest(c *gin.Context) {
	if c.Request.Method == http.MethodOptions {
		handlePreflightRequest(c)
	} else {
		handleProxyRequest(c)
	}
}

func main() {
	const port string = ":8082"
	r := gin.Default()
	r.Any("/proxy", handleRequest)
	fmt.Println("Running liteproxy on", port)
	r.Run(port)
}
