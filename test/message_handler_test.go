package webserver_test

import (
	"testing"
)

// Mock message structures
type Message struct {
	ID             int    `json:"id"`
	SenderID       int    `json:"sender_id"`
	Content        string `json:"content"`
	MessageType    string `json:"message_type"`
	MediaURL       string `json:"media_url,omitempty"`
	MediaType      string `json:"media_type,omitempty"`
	SenderUsername string `json:"sender_username"`
}

type SendMessageRequest struct {
	Content     string `json:"content"`
	MessageType string `json:"message_type"` // "direct" or "broadcast"
	Recipients  []int  `json:"recipients"`   // For direct messages
	MediaURL    string `json:"media_url,omitempty"`
	MediaType   string `json:"media_type,omitempty"`
}

func validateMessageRequest(req SendMessageRequest) []string {
	var errors []string

	// Check message type
	if req.MessageType != "direct" && req.MessageType != "broadcast" {
		errors = append(errors, "Message type must be 'direct' or 'broadcast'")
	}

	// Check content
	if len(req.Content) == 0 && len(req.MediaURL) == 0 {
		errors = append(errors, "Message content or media is required")
	}

	if len(req.Content) > 1000 {
		errors = append(errors, "Message content too long (max 1000 characters)")
	}

	// Check recipients for direct messages
	if req.MessageType == "direct" && len(req.Recipients) == 0 {
		errors = append(errors, "Recipients required for direct messages")
	}

	// Check media type if media URL is provided
	if len(req.MediaURL) > 0 {
		validMediaTypes := map[string]bool{
			"image": true,
			"video": true,
			"audio": true,
			"file":  true,
		}
		if !validMediaTypes[req.MediaType] {
			errors = append(errors, "Invalid media type")
		}
	}

	return errors
}

func TestMessageValidation(t *testing.T) {
	tests := []struct {
		name     string
		request  SendMessageRequest
		hasError bool
	}{
		{
			name: "Valid Direct Message",
			request: SendMessageRequest{
				Content:     "Hello, how are you?",
				MessageType: "direct",
				Recipients:  []int{2, 3},
			},
			hasError: false,
		},
		{
			name: "Valid Broadcast Message",
			request: SendMessageRequest{
				Content:     "Hello everyone!",
				MessageType: "broadcast",
			},
			hasError: false,
		},
		{
			name: "Valid Media Message",
			request: SendMessageRequest{
				Content:     "Check this image",
				MessageType: "direct",
				Recipients:  []int{2},
				MediaURL:    "http://example.com/image.jpg",
				MediaType:   "image",
			},
			hasError: false,
		},
		{
			name: "Invalid Message Type",
			request: SendMessageRequest{
				Content:     "Hello",
				MessageType: "invalid",
				Recipients:  []int{2},
			},
			hasError: true,
		},
		{
			name: "Empty Content and No Media",
			request: SendMessageRequest{
				Content:     "",
				MessageType: "direct",
				Recipients:  []int{2},
			},
			hasError: true,
		},
		{
			name: "Direct Message Without Recipients",
			request: SendMessageRequest{
				Content:     "Hello",
				MessageType: "direct",
			},
			hasError: true,
		},
		{
			name: "Content Too Long",
			request: SendMessageRequest{
				Content:     string(make([]byte, 1001)), // 1001 characters
				MessageType: "direct",
				Recipients:  []int{2},
			},
			hasError: true,
		},
		{
			name: "Invalid Media Type",
			request: SendMessageRequest{
				Content:     "Check this file",
				MessageType: "direct",
				Recipients:  []int{2},
				MediaURL:    "http://example.com/file.exe",
				MediaType:   "executable",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateMessageRequest(tt.request)
			hasError := len(errors) > 0

			if hasError != tt.hasError {
				t.Errorf("validateMessageRequest() hasError = %v, want %v. Errors: %v", hasError, tt.hasError, errors)
			}
		})
	}
}

func TestMessageContentSanitization(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Normal Text",
			content:  "Hello, how are you?",
			expected: "Hello, how are you?",
		},
		{
			name:     "Text with HTML",
			content:  "<script>alert('xss')</script>Hello",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;Hello",
		},
		{
			name:     "Text with Special Characters",
			content:  "Price: $100 & free shipping!",
			expected: "Price: $100 &amp; free shipping!",
		},
		{
			name:     "Empty Content",
			content:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeContent(tt.content)
			if result != tt.expected {
				t.Errorf("sanitizeContent(%q) = %q, want %q", tt.content, result, tt.expected)
			}
		})
	}
}

// Helper function to sanitize HTML content
func sanitizeContent(content string) string {
	// Simple HTML escaping
	replacements := map[rune]string{
		'<':  "&lt;",
		'>':  "&gt;",
		'&':  "&amp;",
		'"':  "&quot;",
		'\'': "&#39;",
	}

	result := ""
	for _, char := range content {
		if replacement, exists := replacements[char]; exists {
			result += replacement
		} else {
			result += string(char)
		}
	}

	return result
}
