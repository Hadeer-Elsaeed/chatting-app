package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	uploadDir string
}

func NewMediaHandler() *MediaHandler {
	uploadDir := getEnvOrDefault("UPLOAD_DIR", "./uploads")
	
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %v", err))
	}
	
	return &MediaHandler{uploadDir: uploadDir}
}

func (h *MediaHandler) UploadMedia(c *gin.Context) {
	userID, username, _ := GetUserFromContext(c)

	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Error:   "Failed to parse form data",
		})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Error:   "No file uploaded",
		})
		return
	}
	defer file.Close()

	// Validate file size (10MB max)
	if header.Size > 10<<20 {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Error:   "File size must be less than 10MB",
		})
		return
	}

	allowedTypes := map[string]bool{
		"image/jpeg":    true,
		"image/jpg":     true,
		"image/png":     true,
		"image/gif":     true,
		"image/webp":    true,
		"video/mp4":     true,
		"video/webm":    true,
		"video/ogg":     true,
		"audio/mp3":     true,
		"audio/wav":     true,
		"audio/ogg":     true,
		"application/pdf": true,
		"text/plain":    true,
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Error:   "File type not allowed",
		})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%s_%d%s", userID, username, time.Now().Unix(), ext)
	
	userDir := filepath.Join(h.uploadDir, fmt.Sprintf("user_%d", userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to create user directory",
		})
		return
	}

	// Save file
	filePath := filepath.Join(userDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to create file",
		})
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to save file",
		})
		return
	}

	// Generate URL
	fileURL := fmt.Sprintf("/api/media/user_%d/%s", userID, filename)

	mediaType := "file"
	if strings.HasPrefix(contentType, "image/") {
		mediaType = "image"
	} else if strings.HasPrefix(contentType, "video/") {
		mediaType = "video"
	} else if strings.HasPrefix(contentType, "audio/") {
		mediaType = "audio"
	}

	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Message: "File uploaded successfully",
		Data: gin.H{
			"url":        fileURL,
			"filename":   header.Filename,
			"size":       header.Size,
			"type":       contentType,
			"media_type": mediaType,
		},
	})
}

func (h *MediaHandler) ServeMedia(c *gin.Context) {
	userDir := c.Param("user_dir")
	filename := c.Param("filename")

	// Validate user directory format
	if !strings.HasPrefix(userDir, "user_") {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Error:   "Invalid user directory",
		})
		return
	}

	filePath := filepath.Join(h.uploadDir, userDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, ApiResponse{
			Success: false,
			Error:   "File not found",
		})
		return
	}

	c.File(filePath)
}

func (h *MediaHandler) GetUserMedia(c *gin.Context) {
	userID, _, _ := GetUserFromContext(c)
	
	userDir := filepath.Join(h.uploadDir, fmt.Sprintf("user_%d", userID))
	
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		c.JSON(http.StatusOK, ApiResponse{
			Success: true,
			Data:    []gin.H{},
		})
		return
	}

	files, err := os.ReadDir(userDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to read media directory",
		})
		return
	}

	var mediaFiles []gin.H
	for _, file := range files {
		if !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				continue
			}

			mediaFiles = append(mediaFiles, gin.H{
				"filename":   file.Name(),
				"size":       info.Size(),
				"created_at": info.ModTime(),
				"url":        fmt.Sprintf("/api/media/user_%d/%s", userID, file.Name()),
			})
		}
	}

	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Data:    mediaFiles,
	})
}
