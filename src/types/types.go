package types

type Project struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Customer Customer `json:"customer"`
	Tasks    []Task   `json:"tasks"`
}

type Customer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TimeEntry struct {
	ID          int     `json:"id"`
	Date        string  `json:"date"`
	Hours       float64 `json:"hours"`
	ProjectID   int     `json:"project_id"`
	TaskID      int     `json:"task_id"`
	Description string  `json:"description"`
	Task        Task    `json:"task"`
}
