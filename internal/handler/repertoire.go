package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pieceResponse struct {
	ID         int     `json:"id"`
	Title      string  `json:"title"`
	Composer   string  `json:"composer"`
	Category   string  `json:"category"`
	Difficulty int     `json:"difficulty"`
	Phase      int     `json:"phase"`
	XPReward   int     `json:"xp_reward"`
	IMSLPUrl   *string `json:"imslp_url,omitempty"`
	SortOrder  int     `json:"sort_order"`
	Status     string  `json:"status"`
}

type updateProgressRequest struct {
	Status string `json:"status" binding:"required,oneof=not_started learning mastering conquered"`
}

func ListPieces(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		category := c.Query("category")

		query := `
			SELECT p.id, p.title, p.composer, p.category, p.difficulty, p.phase,
			       p.xp_reward, p.imslp_url, p.sort_order,
			       COALESCE(pp.status, 'not_started') AS status
			FROM pieces p
			LEFT JOIN piece_progress pp ON pp.piece_id = p.id AND pp.user_id = $1
		`
		args := []interface{}{userID}

		if category != "" {
			query += ` WHERE p.category = $2`
			args = append(args, category)
		}
		query += ` ORDER BY p.sort_order, p.id`

		rows, err := pool.Query(context.Background(), query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		defer rows.Close()

		var pieces []pieceResponse
		for rows.Next() {
			var p pieceResponse
			if err := rows.Scan(
				&p.ID, &p.Title, &p.Composer, &p.Category, &p.Difficulty, &p.Phase,
				&p.XPReward, &p.IMSLPUrl, &p.SortOrder, &p.Status,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error"})
				return
			}
			pieces = append(pieces, p)
		}
		if pieces == nil {
			pieces = []pieceResponse{}
		}

		c.JSON(http.StatusOK, pieces)
	}
}

func UpdateProgress(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		pieceIDStr := c.Param("id")
		pieceID, err := strconv.Atoi(pieceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid piece id"})
			return
		}

		var req updateProgressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify piece exists
		var count int
		err = pool.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM pieces WHERE id = $1`, pieceID,
		).Scan(&count)
		if err != nil || count == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "piece not found"})
			return
		}

		var updatedAt time.Time
		err = pool.QueryRow(context.Background(),
			`INSERT INTO piece_progress (user_id, piece_id, status, updated_at)
			 VALUES ($1, $2, $3, NOW())
			 ON CONFLICT (user_id, piece_id) DO UPDATE
			   SET status = EXCLUDED.status, updated_at = NOW()
			 RETURNING updated_at`,
			userID, pieceID, req.Status,
		).Scan(&updatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update progress"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"piece_id":   pieceID,
			"status":     req.Status,
			"updated_at": updatedAt,
		})
	}
}
