package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type createSessionRequest struct {
	WeekNumber  int    `json:"week_number" binding:"required"`
	DayName     string `json:"day_name" binding:"required"`
	PracticedOn string `json:"practiced_on"` // optional, YYYY-MM-DD; defaults to today
}

type sessionTaskResponse struct {
	ID          string     `json:"id"`
	TaskID      int        `json:"task_id"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type sessionResponse struct {
	ID          string                `json:"id"`
	UserID      string                `json:"user_id"`
	WeekNumber  int                   `json:"week_number"`
	DayName     string                `json:"day_name"`
	PracticedOn string                `json:"practiced_on"`
	CreatedAt   time.Time             `json:"created_at"`
	Tasks       []sessionTaskResponse `json:"tasks"`
}

type streakResponse struct {
	CurrentStreak int    `json:"current_streak"`
	LongestStreak int    `json:"longest_streak"`
	TotalSessions int    `json:"total_sessions"`
	LastPracticed string `json:"last_practiced,omitempty"`
}

func GetOrCreateSession(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		var req createSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		practicedOn := req.PracticedOn
		if practicedOn == "" {
			practicedOn = time.Now().Format("2006-01-02")
		}

		// Upsert session (idempotent by user_id + practiced_on + day_name)
		var sessionID string
		var createdAt time.Time
		err := pool.QueryRow(context.Background(),
			`INSERT INTO practice_sessions (user_id, week_number, day_name, practiced_on)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (user_id, practiced_on, day_name) DO UPDATE
			   SET week_number = EXCLUDED.week_number
			 RETURNING id, created_at`,
			userID, req.WeekNumber, req.DayName, practicedOn,
		).Scan(&sessionID, &createdAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
			return
		}

		// Auto-populate session_tasks from curriculum
		_, err = pool.Exec(context.Background(),
			`INSERT INTO session_tasks (session_id, task_id)
			 SELECT $1, ct.id
			 FROM curriculum_tasks ct
			 JOIN curriculum_days cd ON cd.id = ct.day_id
			 JOIN curriculum_weeks cw ON cw.id = cd.week_id
			 WHERE cw.week_number = $2 AND cd.day_name = $3
			 ON CONFLICT (session_id, task_id) DO NOTHING`,
			sessionID, req.WeekNumber, req.DayName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to populate tasks"})
			return
		}

		tasks, err := fetchSessionTasks(pool, sessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tasks"})
			return
		}

		c.JSON(http.StatusOK, sessionResponse{
			ID:          sessionID,
			UserID:      userID,
			WeekNumber:  req.WeekNumber,
			DayName:     req.DayName,
			PracticedOn: practicedOn,
			CreatedAt:   createdAt,
			Tasks:       tasks,
		})
	}
}

func ToggleTask(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		sessionID := c.Param("id")
		taskID := c.Param("taskId")

		// Verify the session belongs to the user
		var count int
		err := pool.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM practice_sessions WHERE id = $1 AND user_id = $2`,
			sessionID, userID,
		).Scan(&count)
		if err != nil || count == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}

		// Toggle completion
		var completed bool
		var completedAt *time.Time
		err = pool.QueryRow(context.Background(),
			`UPDATE session_tasks
			 SET completed    = NOT completed,
			     completed_at = CASE WHEN NOT completed THEN NOW() ELSE NULL END
			 WHERE session_id = $1 AND task_id = $2
			 RETURNING completed, completed_at`,
			sessionID, taskID,
		).Scan(&completed, &completedAt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found in session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"session_id":   sessionID,
			"task_id":      taskID,
			"completed":    completed,
			"completed_at": completedAt,
		})
	}
}

func GetStreak(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		// Get all distinct practice dates ordered desc
		rows, err := pool.Query(context.Background(),
			`SELECT DISTINCT practiced_on::text
			 FROM practice_sessions
			 WHERE user_id = $1
			 ORDER BY practiced_on DESC`,
			userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		defer rows.Close()

		var dates []time.Time
		for rows.Next() {
			var ds string
			if err := rows.Scan(&ds); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error"})
				return
			}
			t, err := time.Parse("2006-01-02", ds)
			if err != nil {
				continue
			}
			dates = append(dates, t)
		}
		rows.Close()

		totalSessions := len(dates)

		if totalSessions == 0 {
			c.JSON(http.StatusOK, streakResponse{})
			return
		}

		today := time.Now().Truncate(24 * time.Hour)
		lastPracticed := dates[0].Format("2006-01-02")

		// Calculate current streak
		currentStreak := 0
		expected := today
		// if last session was yesterday, still count streak
		if dates[0].Equal(today) || dates[0].Equal(today.AddDate(0, 0, -1)) {
			for _, d := range dates {
				if d.Equal(expected) || d.Equal(expected.AddDate(0, 0, -1)) {
					if d.Equal(expected) {
						currentStreak++
						expected = expected.AddDate(0, 0, -1)
					} else {
						expected = d.AddDate(0, 0, -1)
						currentStreak++
					}
				} else {
					break
				}
			}
		}

		// Recalculate more carefully
		currentStreak = calculateCurrentStreak(dates, today)
		longestStreak := calculateLongestStreak(dates)

		c.JSON(http.StatusOK, streakResponse{
			CurrentStreak: currentStreak,
			LongestStreak: longestStreak,
			TotalSessions: totalSessions,
			LastPracticed: lastPracticed,
		})
	}
}

func calculateCurrentStreak(dates []time.Time, today time.Time) int {
	if len(dates) == 0 {
		return 0
	}
	last := dates[0]
	// streak only counts if practiced today or yesterday
	if !last.Equal(today) && !last.Equal(today.AddDate(0, 0, -1)) {
		return 0
	}
	streak := 1
	for i := 1; i < len(dates); i++ {
		expected := dates[i-1].AddDate(0, 0, -1)
		if dates[i].Equal(expected) {
			streak++
		} else {
			break
		}
	}
	return streak
}

func calculateLongestStreak(dates []time.Time) int {
	if len(dates) == 0 {
		return 0
	}
	longest := 1
	current := 1
	for i := 1; i < len(dates); i++ {
		expected := dates[i-1].AddDate(0, 0, -1)
		if dates[i].Equal(expected) {
			current++
			if current > longest {
				longest = current
			}
		} else {
			current = 1
		}
	}
	return longest
}

func fetchSessionTasks(pool *pgxpool.Pool, sessionID string) ([]sessionTaskResponse, error) {
	rows, err := pool.Query(context.Background(),
		`SELECT id, task_id, completed, completed_at
		 FROM session_tasks
		 WHERE session_id = $1
		 ORDER BY task_id`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []sessionTaskResponse
	for rows.Next() {
		var t sessionTaskResponse
		if err := rows.Scan(&t.ID, &t.TaskID, &t.Completed, &t.CompletedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []sessionTaskResponse{}
	}
	return tasks, nil
}
