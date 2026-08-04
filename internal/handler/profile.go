package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type profileResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	DisplayName  *string   `json:"display_name"`
	CurrentPhase int       `json:"current_phase"`
	CurrentWeek  int       `json:"current_week"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type updateProfileRequest struct {
	DisplayName  *string `json:"display_name"`
	CurrentPhase *int    `json:"current_phase"`
	CurrentWeek  *int    `json:"current_week"`
}

func GetProfile(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		var p profileResponse
		err := pool.QueryRow(context.Background(),
			`SELECT sp.id, sp.user_id, sp.display_name, sp.current_phase, sp.current_week,
			        u.name, u.email, sp.created_at, sp.updated_at
			 FROM student_profiles sp
			 JOIN users u ON u.id = sp.user_id
			 WHERE sp.user_id = $1`,
			userID,
		).Scan(
			&p.ID, &p.UserID, &p.DisplayName, &p.CurrentPhase, &p.CurrentWeek,
			&p.Name, &p.Email, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
			return
		}

		c.JSON(http.StatusOK, p)
	}
}

func UpdateProfile(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		var req updateProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := pool.Exec(context.Background(),
			`UPDATE student_profiles
			 SET display_name   = COALESCE($2, display_name),
			     current_phase  = COALESCE($3, current_phase),
			     current_week   = COALESCE($4, current_week),
			     updated_at     = NOW()
			 WHERE user_id = $1`,
			userID, req.DisplayName, req.CurrentPhase, req.CurrentWeek,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
			return
		}

		var p profileResponse
		err = pool.QueryRow(context.Background(),
			`SELECT sp.id, sp.user_id, sp.display_name, sp.current_phase, sp.current_week,
			        u.name, u.email, sp.created_at, sp.updated_at
			 FROM student_profiles sp
			 JOIN users u ON u.id = sp.user_id
			 WHERE sp.user_id = $1`,
			userID,
		).Scan(
			&p.ID, &p.UserID, &p.DisplayName, &p.CurrentPhase, &p.CurrentWeek,
			&p.Name, &p.Email, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated profile"})
			return
		}

		c.JSON(http.StatusOK, p)
	}
}

func getUserID(c *gin.Context) (string, bool) {
	raw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", false
	}
	id, ok := raw.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id"})
		return "", false
	}
	return id, true
}
