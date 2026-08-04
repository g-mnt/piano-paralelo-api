package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type weekSummary struct {
	ID          int    `json:"id"`
	WeekNumber  int    `json:"week_number"`
	Title       string `json:"title"`
	Theme       string `json:"theme"`
	PhaseNumber int    `json:"phase"`
}

type taskDetail struct {
	ID          int     `json:"id"`
	TaskOrder   int     `json:"task_order"`
	DurationMin int     `json:"duration_min"`
	Category    string  `json:"category"`
	Title       string  `json:"title"`
	Detail      *string `json:"detail,omitempty"`
}

type dayDetail struct {
	ID       int          `json:"id"`
	DayName  string       `json:"day_name"`
	DayOrder int          `json:"day_order"`
	Focus    string       `json:"focus"`
	Tasks    []taskDetail `json:"tasks"`
}

type weekDetail struct {
	ID          int         `json:"id"`
	WeekNumber  int         `json:"week_number"`
	Title       string      `json:"title"`
	Theme       string      `json:"theme"`
	PhaseNumber int         `json:"phase"`
	Days        []dayDetail `json:"days"`
}

func ListWeeks(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(context.Background(),
			`SELECT cw.id, cw.week_number, cw.title, cw.theme, cp.phase_number
			 FROM curriculum_weeks cw
			 JOIN curriculum_phases cp ON cp.id = cw.phase_id
			 ORDER BY cp.phase_number, cw.week_number`,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		defer rows.Close()

		var weeks []weekSummary
		for rows.Next() {
			var w weekSummary
			if err := rows.Scan(&w.ID, &w.WeekNumber, &w.Title, &w.Theme, &w.PhaseNumber); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error"})
				return
			}
			weeks = append(weeks, w)
		}
		if weeks == nil {
			weeks = []weekSummary{}
		}

		c.JSON(http.StatusOK, weeks)
	}
}

func GetWeek(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		n, err := strconv.Atoi(c.Param("n"))
		if err != nil || n < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid week number"})
			return
		}

		var w weekDetail
		err = pool.QueryRow(context.Background(),
			`SELECT cw.id, cw.week_number, cw.title, cw.theme, cp.phase_number
			 FROM curriculum_weeks cw
			 JOIN curriculum_phases cp ON cp.id = cw.phase_id
			 WHERE cw.week_number = $1`,
			n,
		).Scan(&w.ID, &w.WeekNumber, &w.Title, &w.Theme, &w.PhaseNumber)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "week not found"})
			return
		}

		// Fetch days with tasks
		dayRows, err := pool.Query(context.Background(),
			`SELECT id, day_name, day_order, focus
			 FROM curriculum_days
			 WHERE week_id = $1
			 ORDER BY day_order`,
			w.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		defer dayRows.Close()

		var days []dayDetail
		for dayRows.Next() {
			var d dayDetail
			if err := dayRows.Scan(&d.ID, &d.DayName, &d.DayOrder, &d.Focus); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error"})
				return
			}
			days = append(days, d)
		}
		dayRows.Close()

		// Fetch tasks for each day
		for i, d := range days {
			taskRows, err := pool.Query(context.Background(),
				`SELECT id, task_order, duration_min, category, title, detail
				 FROM curriculum_tasks
				 WHERE day_id = $1
				 ORDER BY task_order`,
				d.ID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			var tasks []taskDetail
			for taskRows.Next() {
				var t taskDetail
				if err := taskRows.Scan(&t.ID, &t.TaskOrder, &t.DurationMin, &t.Category, &t.Title, &t.Detail); err != nil {
					taskRows.Close()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error"})
					return
				}
				tasks = append(tasks, t)
			}
			taskRows.Close()

			if tasks == nil {
				tasks = []taskDetail{}
			}
			days[i].Tasks = tasks
		}

		if days == nil {
			days = []dayDetail{}
		}
		w.Days = days

		c.JSON(http.StatusOK, w)
	}
}
