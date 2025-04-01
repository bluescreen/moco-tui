package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
)

type Model struct {
	taskID       string
	taskTitle    string
	projectID    string
	date         string
	hours        string
	description  string
	errorMsg     string
	submitting   bool
	width        int
	height       int
	taskList     list.Model
	dateInput    textinput.Model
	hoursInput   textinput.Model
	descInput    textinput.Model
	focusedInput int
}

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
	Date        string `json:"date"`
	Hours       string `json:"hours"`
	ProjectID   int    `json:"project_id"`
	TaskID      int    `json:"task_id"`
	Description string `json:"description"`
}

type item struct {
	title, desc     string
	taskID          int
	projectID       int
	isProjectHeader bool
	position        int
}

func fetchProjects(mocoDomain, apiKey string) ([]Project, error) {
	url := fmt.Sprintf("https://%s.mocoapp.com/api/v1/projects/assigned", mocoDomain)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func submitTimeEntry(mocoDomain, apiKey string, entry TimeEntry) error {
	url := fmt.Sprintf("https://%s.mocoapp.com/api/v1/time_entries", mocoDomain)
	client := &http.Client{}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s", string(body))
	}

	return nil
}

func (m *Model) receiveTimeEntryInput() {
	m.date = time.Now().Format("2006-01-02")
	m.hours = "0"
	m.description = ""
	m.submitting = true
}

func (m *Model) handleTimeEntrySubmission() {
	projectID, err := strconv.Atoi(fmt.Sprintf("%d", m.projectID))
	taskID, err := strconv.Atoi(fmt.Sprintf("%d", m.taskID))
	entry := TimeEntry{
		Date:        m.date,
		Hours:       m.hours,
		ProjectID:   projectID,
		TaskID:      taskID,
		Description: m.description,
	}
	log.Printf("%+v", entry)

	err = submitTimeEntry(os.Getenv("MOCO_DOMAIN"), os.Getenv("MOCO_API_KEY"), entry)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Error submitting time entry: %v", err)
	} else {
		m.errorMsg = "Time entry submitted successfully!"
	}
	m.submitting = false
}

func (m Model) Init() tea.Cmd {

	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			if m.submitting {
				m.handleTimeEntrySubmission()
			} else {
				selected := m.taskList.SelectedItem()
				if selected == nil {
					return m, nil
				}
				selectedItem, ok := selected.(item)
				if !ok {
					return m, nil
				}
				if !selectedItem.isProjectHeader {
					m.taskID = fmt.Sprintf("%d", selectedItem.taskID)
					m.projectID = fmt.Sprintf("%d", selectedItem.projectID)
					m.taskTitle = selectedItem.desc
					m.receiveTimeEntryInput()
				}
			}
		case "j":
			m.taskList.CursorDown()
		case "k":
			m.taskList.CursorUp()
		case "down":
			if m.submitting {
				m.focusedInput = (m.focusedInput + 1) % 3
				m.focusCurrentInput()
			} else {
				m.taskList.CursorDown()
			}
		case "up":
			if m.submitting {
				m.focusedInput = (m.focusedInput - 1 + 3) % 3
				m.focusCurrentInput()
			} else {
				m.taskList.CursorUp()
			}

		case "tab":
			if m.submitting {
				m.focusedInput = (m.focusedInput + 1) % 3
				m.focusCurrentInput()
			}
		case "shift+tab":
			if m.submitting {
				m.focusedInput = (m.focusedInput - 1 + 3) % 3
				m.focusCurrentInput()
			}
		}

		m.dateInput, _ = m.dateInput.Update(msg)
		m.hoursInput, _ = m.hoursInput.Update(msg)
		m.descInput, _ = m.descInput.Update(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.taskList.SetSize(msg.Width-h, msg.Height-v)

	default:
		log.Printf("Update msg: %+v\n", msg) // ADD THIS LINE
	}

	return m, nil
}

func (m Model) View() string {
	var content string
	if m.errorMsg != "" {
		content = fmt.Sprintf("%s\n\nPress 'q' to quit.\n", m.errorMsg)
	} else if m.submitting {
		content = lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("Book Entry\n"),
			fmt.Sprintf("%s\n", m.taskTitle),
			fmt.Sprintf("Hours: %s", m.hoursInput.View()),
			fmt.Sprintf("Description: %s", m.descInput.View()),
			fmt.Sprintf("Date: %s", m.dateInput.View()),
			"\nPress 'enter' to submit, 'q' to quit.",
		)
	} else {
		content = fmt.Sprintf("Select a Task:\n%s\nUse 'j'/'k' to navigate, 'enter' to select task, 'q' to quit.", m.taskList.View())
	}

	padding := 0
	if m.submitting || m.errorMsg != "" {
		padding = 5
	}
	paneStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(padding/2, padding)

	pane := paneStyle.Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, pane)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	mocoDomain := os.Getenv("MOCO_DOMAIN")
	apiKey := os.Getenv("MOCO_API_KEY")

	if mocoDomain == "" || apiKey == "" {
		log.Fatal("Error: MOCO_DOMAIN and MOCO_API_KEY environment variables must be set.")
	}

	projects, err := fetchProjects(mocoDomain, apiKey)
	if err != nil {
		log.Fatal("Error fetching projects:", err)
	}

	var items []list.Item

	for _, project := range projects {
		i := 1
		if len(project.Tasks) > 0 {
			items = append(items, item{
				title:           project.Name,
				desc:            "Project",
				isProjectHeader: true,
			})

			sort.Slice(project.Tasks, func(i, j int) bool {
				return project.Tasks[i].Name < project.Tasks[j].Name
			})

			for _, task := range project.Tasks {
				items = append(items, item{
					desc:            task.Name,
					title:           project.Customer.Name,
					taskID:          task.ID,
					projectID:       project.ID,
					isProjectHeader: false,
					position:        i,
				})
				i++
			}
			i = 1
		}
	}

	taskList := list.New(items, itemDelegate{}, 0, 0)
	taskList.Title = "MOCO - Select a task:"

	dateInput := textinput.New()
	dateInput.Placeholder = "Enter date (YYYY-MM-DD)"

	hoursInput := textinput.New()
	hoursInput.Placeholder = "Enter hours"

	descInput := textinput.New()
	descInput.Placeholder = "Enter description"

	hoursInput.Focus()

	model := Model{taskList: taskList, dateInput: dateInput, hoursInput: hoursInput, descInput: descInput}

	p := tea.NewProgram(model, tea.WithAltScreen())
	p.Run()
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return "" }

var (
	docStyle          = lipgloss.NewStyle().Margin(1, 2)
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("255"))
)

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	str := ""
	fn := itemStyle.Render

	if i.isProjectHeader {
		str = fmt.Sprintf("%s", i.title)
	} else {
		str = fmt.Sprintf("\t [%d] %s", i.position, i.desc)

		if index == m.Index() {
			fn = func(s ...string) string {
				return selectedItemStyle.Render("> " + strings.Join(s, " "))
			}
		}
	}

	fmt.Fprint(w, fn(str))
}

func (m *Model) focusCurrentInput() {
	m.dateInput.Blur()
	m.hoursInput.Blur()
	m.descInput.Blur()

	switch m.focusedInput {
	case 0:
		m.dateInput.Focus()
	case 1:
		m.hoursInput.Focus()
	case 2:
		m.descInput.Focus()
	}
}
