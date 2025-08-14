package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	db *sql.DB
}

func NewMessageHandler(db *sql.DB) *MessageHandler {
	return &MessageHandler{db: db}
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	senderID, _, _ := GetUserFromContext(c)

	if req.MessageType != "direct" && req.MessageType != "broadcast" {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Error:   "Invalid message type. Must be 'direct' or 'broadcast'",
		})
		return
	}

	if req.MessageType == "direct" && len(req.Recipients) == 0 {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Error:   "Recipients required for direct messages",
		})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		"INSERT INTO messages (sender_id, content, message_type, media_url, media_type) VALUES (?, ?, ?, ?, ?)",
		senderID, req.Content, req.MessageType, req.MediaURL, req.MediaType,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to create message",
		})
		return
	}

	messageID, _ := result.LastInsertId()

	if req.MessageType == "direct" {
		for _, recipientID := range req.Recipients {
			_, err := tx.Exec(
				"INSERT INTO message_recipients (message_id, recipient_id) VALUES (?, ?)",
				messageID, recipientID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ApiResponse{
					Success: false,
					Error:   "Failed to add recipients",
				})
				return
			}
		}
	} else {
		_, err := tx.Exec(
			"INSERT INTO message_recipients (message_id, recipient_id) SELECT ?, id FROM users WHERE id != ?",
			messageID, senderID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ApiResponse{
				Success: false,
				Error:   "Failed to add broadcast recipients",
			})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to send message",
		})
		return
	}

	// Get the created message with sender info
	var message Message
	err = h.db.QueryRow(`
		SELECT m.id, m.sender_id, m.content, m.message_type, m.media_url, m.media_type, m.created_at, u.username
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.id = ?
	`, messageID).Scan(
		&message.ID, &message.SenderID, &message.Content, &message.MessageType,
		&message.MediaURL, &message.MediaType, &message.CreatedAt, &message.SenderUsername,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Message sent but failed to retrieve details",
		})
		return
	}

	// Notify WebSocket server about the new message
	var recipientIDs []int
	if req.MessageType == "direct" {
		recipientIDs = req.Recipients
	} else {
		rows, err := h.db.Query("SELECT id FROM users WHERE id != ?", senderID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var userID int
				if err := rows.Scan(&userID); err == nil {
					recipientIDs = append(recipientIDs, userID)
				}
			}
		}
	}

	go NotifyWebSocketServer(message, recipientIDs)

	c.JSON(http.StatusCreated, ApiResponse{
		Success: true,
		Message: "Message sent successfully",
		Data:    message,
	})
}

func (h *MessageHandler) GetMessageHistory(c *gin.Context) {
	userID, _, _ := GetUserFromContext(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	messageType := c.Query("type") // "direct", "broadcast", or empty for all

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	query := `
		SELECT DISTINCT m.id, m.sender_id, m.content, m.message_type, m.media_url, m.media_type, m.created_at, u.username
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		LEFT JOIN message_recipients mr ON m.id = mr.message_id
		WHERE (mr.recipient_id = ? OR m.sender_id = ?)
	`

	args := []interface{}{userID, userID}

	if messageType != "" {
		query += " AND m.message_type = ?"
		args = append(args, messageType)
	}

	query += " ORDER BY m.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to fetch message history",
		})
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var message Message
		err := rows.Scan(
			&message.ID, &message.SenderID, &message.Content, &message.MessageType,
			&message.MediaURL, &message.MediaType, &message.CreatedAt, &message.SenderUsername,
		)
		if err != nil {
			continue
		}
		messages = append(messages, message)
	}

	countQuery := `
		SELECT COUNT(DISTINCT m.id)
		FROM messages m
		LEFT JOIN message_recipients mr ON m.id = mr.message_id
		WHERE (mr.recipient_id = ? OR m.sender_id = ?)
	`

	countArgs := []interface{}{userID, userID}

	if messageType != "" {
		countQuery += " AND m.message_type = ?"
		countArgs = append(countArgs, messageType)
	}

	var total int
	err = h.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		total = 0
	}

	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Data: MessageHistoryResponse{
			Messages: messages,
			Total:    total,
			Page:     page,
			Limit:    limit,
		},
	})
}

func (h *MessageHandler) GetConversation(c *gin.Context) {
	userID, _, _ := GetUserFromContext(c)
	otherUserIDStr := c.Param("user_id")

	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	query := `
		SELECT m.id, m.sender_id, m.content, m.message_type, m.media_url, m.media_type, m.created_at, u.username
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		JOIN message_recipients mr ON m.id = mr.message_id
		WHERE m.message_type = 'direct' 
		AND (
			(m.sender_id = ? AND mr.recipient_id = ?) OR 
			(m.sender_id = ? AND mr.recipient_id = ?)
		)
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := h.db.Query(query, userID, otherUserID, otherUserID, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, ApiResponse{
			Success: true,
			Data:    []Message{},
			Message: "No Conversation found",
		})
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var message Message
		err := rows.Scan(
			&message.ID, &message.SenderID, &message.Content, &message.MessageType,
			&message.MediaURL, &message.MediaType, &message.CreatedAt, &message.SenderUsername,
		)
		if err != nil {
			continue
		}
		messages = append(messages, message)
	}

	// order to show oldest first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Data:    messages,
	})
}

func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	userID, _, _ := GetUserFromContext(c)
	messageIDsStr := c.Query("message_ids")

	if messageIDsStr == "" {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Error:   "Message IDs required",
		})
		return
	}

	messageIDStrs := strings.Split(messageIDsStr, ",")
	messageIDs := make([]interface{}, len(messageIDStrs))

	for i, idStr := range messageIDStrs {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			c.JSON(http.StatusBadRequest, ApiResponse{
				Success: false,
				Error:   "Invalid message ID: " + idStr,
			})
			return
		}
		messageIDs[i] = id
	}

	placeholders := strings.Repeat("?,", len(messageIDs)-1) + "?"
	query := `UPDATE message_recipients 
			  SET is_read = true, read_at = NOW() 
			  WHERE recipient_id = ? AND message_id IN (` + placeholders + `)`

	args := append([]interface{}{userID}, messageIDs...)

	_, err := h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to mark messages as read",
		})
		return
	}

	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Message: "Messages marked as read",
	})
}
