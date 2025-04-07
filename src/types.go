package main

type Project struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Tasks    []Task   `json:"tasks"`
	Customer Customer `json:"customer"`
}

type Customer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Task struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Active   bool   `json:"active"`
	Billable bool   `json:"billable"`
}

type TimeEntry struct {
	ID          int     `json:"id"`
	Date        string  `json:"date"`
	Hours       float64 `json:"hours"`
	ProjectID   int     `json:"project_id"`
	TaskID      int     `json:"task_id"`
	Description string  `json:"description"`
	UserID      int     `json:"user_id"`
	Billable    bool    `json:"billable"`
	Locked      bool    `json:"locked"`
}

type item struct {
	title, desc     string
	taskID          int
	projectID       int
	isProjectHeader bool
	position        int
}
