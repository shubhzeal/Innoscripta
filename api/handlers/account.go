package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	db *sql.DB
}

func NewAccountHandler(db *sql.DB) *AccountHandler {
	return &AccountHandler{db: db}
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req struct {
		Name    string  `json:"name" binding:"required"`
		Balance float64 `json:"balance" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `INSERT INTO accounts (name, balance) VALUES ($1, $2) RETURNING id`
	var accountID int
	err := h.db.QueryRow(query, req.Name, req.Balance).Scan(&accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"account_id": accountID})
}
