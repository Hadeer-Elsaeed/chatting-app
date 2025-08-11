package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ? OR email = ?)", 
		req.Username, req.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Database error",
		})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, ApiResponse{
			Success: false,
			Error:   "Username or email already exists",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to hash password",
		})
		return
	}

	result, err := h.db.Exec(
		"INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
		req.Username, req.Email, string(hashedPassword),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to create user",
		})
		return
	}

	userID, _ := result.LastInsertId()

	user := User{
		ID:       int(userID),
		Username: req.Username,
		Email:    req.Email,
	}

	token, err := GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusCreated, ApiResponse{
		Success: true,
		Message: "User registered successfully",
		Data: AuthResponse{
			Token: token,
			User:  user,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var user User
	var hashedPassword string
	err := h.db.QueryRow(
		"SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.Email, &hashedPassword, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, ApiResponse{
			Success: false,
			Error:   "Invalid credentials",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Database error",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ApiResponse{
			Success: false,
			Error:   "Invalid credentials",
		})
		return
	}

	token, err := GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Message: "Login successful",
		Data: AuthResponse{
			Token: token,
			User:  user,
		},
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, _, _ := GetUserFromContext(c)

	var user User
	err := h.db.QueryRow(
		"SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, ApiResponse{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Data:    user,
	})
}

func (h *AuthHandler) GetUsers(c *gin.Context) {
	userID, _, _ := GetUserFromContext(c)

	rows, err := h.db.Query(
		"SELECT id, username, email, created_at FROM users WHERE id != ? ORDER BY username",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Error:   "Failed to fetch users",
		})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Data:    UserListResponse{Users: users},
	})
}
