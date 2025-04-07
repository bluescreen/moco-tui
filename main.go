package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/denwerk/moco/src/ui"
	"github.com/joho/godotenv"
)

type Model struct {
	taskID           string
	taskTitle        string
	projectID        string
	timeEntries      []TimeEntry
	date             string
	hours            string
	description      string
	errorMsg         string
	succesMsg        string
	submitting       bool
	width            int
	height           int
	taskList         list.Model
	dateInput        textinput.Model
	hoursInput       textinput.Model
	descInput        textinput.Model
	timeEntriesTable table.Model
	focusedInput     int
	focusedPane      string
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
	hours, err := strconv.ParseFloat(m.hours, 64)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Error parsing hours: %v", err)
		return
	}
	entry := TimeEntry{
		Date:        m.date,
		Hours:       hours,
		ProjectID:   projectID,
		TaskID:      taskID,
		Description: m.description,
	}
	log.Printf("%+v", entry)

	err = submitTimeEntry(os.Getenv("MOCO_DOMAIN"), os.Getenv("MOCO_API_KEY"), entry)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Error submitting time entry: %v", err)
	} else {
		m.succesMsg = "Time entry submitted successfully!"
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

		case "esc":
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

func (m *Model) View() string {
	// Left Pane
	leftPane := m.taskList.View()
	leftPaneStyle := ui.PaneStyle.Width(m.width/2 - 2).Height(m.height - 2)
	if m.focusedPane == "left" {
		leftPaneStyle = ui.FocusedPaneStyle.Width(m.width/2 - 2).Height(m.height - 2)
	}
	leftPane = leftPaneStyle.Render(leftPane)

	// Right Pane
	rightPaneTitle := ui.TitleStyle.Render("Today's Time Entries")
	rightPaneStyle := ui.PaneStyle.Width(m.width/2 - 2).Height(m.height - 2)
	if m.focusedPane == "right" {
		rightPaneStyle = ui.FocusedPaneStyle.Width(m.width/2 - 2).Height(m.height - 2)
	}

	var rightPaneContent string
	if m.errorMsg != "" {
		rightPaneContent = ui.ErrorStyle.Render(m.errorMsg)
	} else if m.succesMsg != "" {
		rightPaneContent = ui.SuccessStyle.Render(m.succesMsg)
	} else {
		rightPaneContent = m.timeEntriesTable.View()
	}

	rightPane := lipgloss.JoinVertical(lipgloss.Left, rightPaneTitle, rightPaneContent)
	rightPane = rightPaneStyle.Render(rightPane)

	// Layout
	layout := lipgloss.JoinHorizontal(lipgloss.Left, leftPane, rightPane)
	return ui.DocStyle.Render(layout)
}

func (m *Model) loadTimeEntries() {
	entries, err := fetchTimeEntries(os.Getenv("MOCO_DOMAIN"), os.Getenv("MOCO_API_KEY"), time.Now().Format("2006-01-02"))
	if err != nil {
		m.errorMsg = fmt.Sprintf("Error loading time entries: %v", err)
	} else {
		m.timeEntries = entries
		m.updateTable()
	}
}

func (m *Model) updateTable() {
	columns := []table.Column{
		{Title: "Description", Width: 30},
		{Title: "Hours", Width: 10},
		{Title: "Task ID", Width: 10},
	}

	var rows []table.Row
	for _, entry := range m.timeEntries {
		rows = append(rows, table.Row{entry.Description, fmt.Sprintf("%.2f", entry.Hours), string(entry.TaskID)})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		BorderBottomForeground(lipgloss.Color("56"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(s)
	m.timeEntriesTable = t
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

	entries, err := fetchTimeEntries(os.Getenv("MOCO_DOMAIN"), os.Getenv("MOCO_API_KEY"), time.Now().Format("2006-01-02"))
	if err != nil {
		log.Fatal("Error fetching time entries:", err)
	}

	fmt.Sprintf("%+v", entries)

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

	// model.timeEntries = entries
	// model.loadTimeEntries()

	p := tea.NewProgram(model, tea.WithAltScreen())
	p.Run()
}

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

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return "" }

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
