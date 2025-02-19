package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"liteproxy/src/redis"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func addCorsHeaders(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
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

	val, err := redis.GetRequest(c.Query("url"))
	if err == nil && val != "" {
		var cachedResponse interface{}
		err := json.Unmarshal([]byte(val), &cachedResponse)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse cached response"})
			return
		}
		c.JSON(http.StatusOK, cachedResponse)
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

	// Copy headers from response
	exposedHeaders := map[string]bool{
		"accept-ranges":       true,
		"age":                 true,
		"cache-control":       true,
		"content-length":      true,
		"content-language":    true,
		"content-type":        true,
		"date":                true,
		"etag":                true,
		"expires":             true,
		"last-modified":       true,
		"location":            true,
		"pragma":              true,
		"server":              true,
		"transfer-encoding":   true,
		"vary":                true,
		"x-github-request-id": true,
		"x-redirected-url":    true,
	}
	for key, values := range resp.Header {
		_, exists := exposedHeaders[strings.ToLower(key)]
		if key == "Content-Length" || key == "Transfer-Encoding" || key == "Connection" || exists == false {
			continue
		}
		for _, value := range values {
			c.Header(key, value)
			// c.Writer.Header().Add(key, value)
		}
	}

	fmt.Println(resp.Header)

	encoding := resp.Header.Get("Content-Encoding")
	if encoding != "" {
		resp.Body, err = decompressBody(resp.Body, encoding)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decompress response"})
			return
		}
	}

	// respBody, err := io.ReadAll(resp.Body) // Read the response body into memory
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
	// 	return
	// }

	// redis.SetRequest(c.Query("url"), string(respBody))

	addCorsHeaders(c)
	c.Status(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

func setGinMode() {
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == gin.DebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

func main() {
	const port string = ":8082"
	r := gin.Default()
	setGinMode()
	r.OPTIONS("/proxy", handlePreflightRequest)
	r.GET("/proxy", handleProxyRequest)
	fmt.Println("Running liteproxy on", port)
	r.Run(port)
}
