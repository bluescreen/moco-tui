package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/denwerk/moco/src/types"
	"github.com/denwerk/moco/src/ui"
	"github.com/joho/godotenv"
)

type Model struct {
	cfg              *Config
	taskID           string
	taskTitle        string
	projectID        string
	timeEntries      []types.TimeEntry
	errorMsg         string
	succesMsg        string
	width            int
	height           int
	taskList         list.Model
	timeEntriesTable table.Model
	ticker           *time.Ticker
	focusedPane      string           // "left", "form", or "timeEntries"
	confirmDelete    bool             // Whether to show delete confirmation
	selectedEntry    *types.TimeEntry // Currently selected time entry
	lastUpdate       time.Time        // When time entries were last updated
	form             ui.FormEntry
	messageTimer     *time.Timer // Timer for clearing messages
}

// parseHours converts a string to hours, supporting both decimal and time format
func parseHours(input string) (float64, error) {
	// Try decimal format first (e.g. "1.5")
	if hours, err := strconv.ParseFloat(input, 64); err == nil {
		if hours <= 0 {
			return 0, fmt.Errorf("hours must be greater than 0")
		}
		return hours, nil
	}

	// Try time format (e.g. "1:30")
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid format. use decimal (e.g. 1.5) or time (e.g. 1:30)")
	}

	fmt.Println(parts)

	hours, err1 := strconv.ParseFloat(parts[0], 64)
	minutes, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return 0, fmt.Errorf("invalid format. use decimal (e.g. 1.5) or time (e.g. 1:30)")
	}

	if minutes < 0 || minutes >= 60 {
		return 0, fmt.Errorf("minutes must be between 0 and 59")
	}

	totalHours := hours + (minutes / 60.0)
	if totalHours <= 0 {
		return 0, fmt.Errorf("hours must be greater than 0")
	}

	return totalHours, nil
}

func (m *Model) handleTimeEntrySubmission() {
	// Clear previous messages
	m.errorMsg = ""
	m.succesMsg = ""

	// Validate project selection
	if m.projectID == "" || m.taskID == "" {
		m.setMessage("Please select a project first", true)
		return
	}

	date, hours, description := m.form.GetValues()

	// Validate hours
	hoursFloat, err := parseHours(hours)
	if err != nil {
		m.setMessage(err.Error(), true)
		return
	}

	// Validate date
	if date == "" {
		m.setMessage("Date is required", true)
		return
	}

	// Validate description
	if description == "" {
		m.setMessage("Description is required", true)
		return
	}

	projectID, err := strconv.Atoi(m.projectID)
	if err != nil {
		m.setMessage("Invalid project ID", true)
		return
	}

	taskID, err := strconv.Atoi(m.taskID)
	if err != nil {
		m.setMessage("Invalid task ID", true)
		return
	}

	entry := types.TimeEntry{
		Date:        date,
		Hours:       hoursFloat,
		ProjectID:   projectID,
		TaskID:      taskID,
		Description: description,
	}

	err = submitTimeEntry(m.cfg, entry)
	if err != nil {
		m.setMessage(fmt.Sprintf("Error submitting time entry: %v", err), true)
	} else {
		m.setMessage("Time entry submitted successfully!", false)
		m.form.Clear()
		m.loadTimeEntries()
	}
}

func (m *Model) handleDeleteTimeEntry() {
	if m.selectedEntry == nil {
		return
	}

	err := deleteTimeEntry(m.cfg, m.selectedEntry.ID)
	if err != nil {
		m.setMessage(fmt.Sprintf("Error deleting time entry: %v", err), true)
	} else {
		m.setMessage("Time entry deleted successfully!", false)
		m.loadTimeEntries()
	}
	m.confirmDelete = false
	m.selectedEntry = nil
}

func (m *Model) Init() tea.Cmd {
	// Start the ticker for polling time entries
	m.ticker = time.NewTicker(10 * time.Second)
	m.focusedPane = "form" // Start with focus on form pane

	// Return a command that will be executed immediately
	return tea.Batch(
		// Initial load of time entries
		func() tea.Msg {
			m.loadTimeEntries()
			return nil
		},
		// Start the ticker
		func() tea.Msg {
			for range m.ticker.C {
				return tea.Msg("tick")
			}
			return nil
		},
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd = m.handleKeyMsg(msg)
	case tea.MouseMsg:
		cmd = m.handleMouseMsg(msg)
	case tea.WindowSizeMsg:
		m.handleWindowSizeMsg(msg)
	case string:
		if msg == "tick" {
			m.loadTimeEntries()
			cmd = m.tickerCmd()
		}
	}

	// Check if message timer has expired
	if m.messageTimer != nil {
		select {
		case <-m.messageTimer.C:
			m.errorMsg = ""
			m.succesMsg = ""
			m.messageTimer = nil
			// Return a command to force a screen update
			return m, tea.Tick(time.Millisecond, func(time.Time) tea.Msg {
				return nil
			})
		default:
		}
	}

	return m, cmd
}

func (m *Model) handleMouseMsg(msg tea.MouseMsg) tea.Cmd {
	if msg.Type != tea.MouseLeft {
		return nil
	}

	leftWidth := m.width / 2
	formHeight := 10

	if msg.X < leftWidth {
		m.focusedPane = "left"
	} else if msg.X >= leftWidth && msg.Y < formHeight {
		m.focusedPane = "form"
	} else if msg.X >= leftWidth && msg.Y >= formHeight {
		m.focusedPane = "timeEntries"
	}

	return nil
}

func (m *Model) handleWindowSizeMsg(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	h, v := ui.DocStyle.GetFrameSize()
	m.taskList.SetSize(msg.Width/2-h, msg.Height-v)
	m.timeEntriesTable.SetWidth(msg.Width/2 - h)
}

func (m *Model) updateTaskInfo() {
	selected := m.taskList.SelectedItem()
	if selected == nil {
		return
	}

	selectedItem, ok := selected.(ui.TableEntry)
	if !ok || selectedItem.IsProjectHeader {
		return
	}

	m.taskID = fmt.Sprintf("%d", selectedItem.TaskID)
	m.projectID = fmt.Sprintf("%d", selectedItem.ProjectID)
	m.taskTitle = selectedItem.Desc
	m.form.SetTaskTitle(m.taskTitle)
}

func (m *Model) updateSelectedEntry() {
	cursor := m.timeEntriesTable.Cursor()
	if cursor > 0 {
		for i := range m.timeEntries {
			if i == cursor {
				m.selectedEntry = &m.timeEntries[i]
				break
			}
		}
	} else {
		m.selectedEntry = nil
	}
}

func (m *Model) blurAllInputs() {
	m.form.BlurAll()
}

func (m *Model) tickerCmd() tea.Cmd {
	return func() tea.Msg {
		for range m.ticker.C {
			return tea.Msg("tick")
		}
		return nil
	}
}

func (m Model) View() string {
	// Calculate pane widths
	leftWidth := m.width / 2
	rightWidth := (m.width / 2) - 5

	// Left Pane (Tasks)
	leftPane := fmt.Sprintf("%s", m.taskList.View())
	leftPaneStyle := ui.PaneStyle.Width(leftWidth).Height(m.height - 2) // Subtract 2 for margins
	if m.focusedPane == "left" {
		leftPaneStyle = ui.FocusedPaneStyle.Width(leftWidth).Height(m.height - 2)
	}
	leftPane = leftPaneStyle.Render(leftPane)

	// Right Pane (Form and Time Entries)
	// Form Section
	formContent := lipgloss.JoinVertical(lipgloss.Left,
		m.form.View(),
	)

	// Add error message if present
	if m.errorMsg != "" {
		formContent = lipgloss.JoinVertical(lipgloss.Left,
			formContent,
			ui.ErrorStyle.Render(fmt.Sprintf("\nError: %s", m.errorMsg)),
		)
	}

	// Add success message if present
	if m.succesMsg != "" {
		formContent = lipgloss.JoinVertical(lipgloss.Left,
			formContent,
			ui.SuccessStyle.Render(fmt.Sprintf("\nSuccess: %s", m.succesMsg)),
		)
	}

	formStyle := ui.PaneStyle.Width(rightWidth)
	if m.focusedPane == "form" {
		formStyle = ui.FocusedPaneStyle.Width(rightWidth)
	}

	formPane := formStyle.Render(formContent)

	// Time Entries Section
	timeEntriesTitle := ui.TitleStyle.Render("Time Entries")
	lastUpdate := ui.LastUpdateStyle.Render(fmt.Sprintf("Last updated: %s", m.lastUpdate.Format("15:04:05")))

	// Add selected entry ID to header if one is selected
	selectedInfo := ""
	if m.selectedEntry != nil {
		selectedInfo = fmt.Sprintf(" (Selected: #%d)", m.selectedEntry.ID)
	}
	header := lipgloss.JoinVertical(lipgloss.Left,
		timeEntriesTitle,
		lastUpdate,
		ui.SelectedStyle.Render(selectedInfo),
	)

	timeEntriesContent := lipgloss.JoinVertical(lipgloss.Left,
		header,
		m.timeEntriesTable.View(),
	)

	timeEntriesStyle := ui.PaneStyle.Width(rightWidth).Height(m.height/2 - 2) // Make time entries pane take up half the height
	if m.focusedPane == "timeEntries" {
		timeEntriesStyle = ui.FocusedPaneStyle.Width(rightWidth).Height(m.height/2 - 2)
	}

	timeEntriesPane := timeEntriesStyle.Render(timeEntriesContent)

	// Combine form and time entries vertically
	rightPane := lipgloss.JoinVertical(lipgloss.Left,
		formPane,
		timeEntriesPane,
	)

	// Add delete confirmation dialog if needed
	if m.confirmDelete && m.selectedEntry != nil {
		confirmDialog := ui.ConfirmDialogStyle.Width(rightWidth).
			Render(fmt.Sprintf(
				"Are you sure you want to delete this time entry?\n"+
					"Date: %s\n"+
					"Hours: %.2f\n"+
					"Description: %s\n\n"+
					"Press ENTER to confirm, ESC to cancel",
				m.selectedEntry.Date,
				m.selectedEntry.Hours,
				m.selectedEntry.Description,
			))
		rightPane = lipgloss.JoinVertical(lipgloss.Left, confirmDialog, rightPane)
	}

	// Layout
	layout := lipgloss.JoinHorizontal(lipgloss.Left, leftPane, rightPane)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, layout)
}

func (m *Model) loadTimeEntries() {
	entries, err := fetchTimeEntries(m.cfg, time.Now().Format("2006-01-02"))
	if err != nil {
		m.errorMsg = fmt.Sprintf("Error loading time entries: %v", err)
	} else {
		m.timeEntries = entries
		m.lastUpdate = time.Now()
		m.updateTable()
	}
}

func (m *Model) updateTable() {
	m.timeEntriesTable = ui.CreateTimeEntriesTable(m.timeEntries, 20)

	// Set the selected entry based on the current selection
	if row := m.timeEntriesTable.SelectedRow(); len(row) > 0 {
		for i, entry := range m.timeEntries {
			if fmt.Sprintf("%d", entry.ID) == row[0] {
				m.selectedEntry = &m.timeEntries[i]
				break
			}
		}
	} else {
		m.selectedEntry = nil
	}
}

func (m *Model) saveLastTask() {
	if m.projectID == "" || m.taskID == "" {
		return
	}

	projectID, err := strconv.Atoi(m.projectID)
	if err != nil {
		return
	}

	taskID, err := strconv.Atoi(m.taskID)
	if err != nil {
		return
	}

	SaveLastTask(LastTask{
		ProjectID: projectID,
		TaskID:    taskID,
		TaskTitle: m.taskTitle,
	})
}

func (m *Model) loadLastTask() {
	lastTask, err := LoadLastTask()
	if err != nil {
		log.Printf("Error loading last task: %v", err)
		return
	}

	if lastTask != nil {
		m.projectID = fmt.Sprintf("%d", lastTask.ProjectID)
		m.taskID = fmt.Sprintf("%d", lastTask.TaskID)
		m.taskTitle = lastTask.TaskTitle
		m.form.SetTaskTitle(m.taskTitle)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	if err := InitLogger(); err != nil {
		log.Fatal("Error initializing logger:", err)
	}
	defer Close()

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	projects, err := fetchProjects(cfg)
	if err != nil {
		log.Fatal("Error fetching projects:", err)
	}

	model := newModel(cfg, projects)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	p.Run()
}

func newModel(cfg *Config, projects []types.Project) *Model {
	items := ui.MapProjectsToItems(projects)
	taskList := list.New(items, ui.ItemDelegate{}, 0, 0)
	taskList.Title = "MOCO " + cfg.MocoDomain + " - Select a task:"

	model := &Model{
		cfg:      cfg,
		taskList: taskList,
		form:     ui.NewFormEntry(),
	}

	model.loadLastTask()
	model.selectLastTask(items)
	model.loadTimeEntries()

	return model
}

func (m *Model) selectLastTask(items []list.Item) {
	if m.projectID == "" || m.taskID == "" {
		return
	}

	for i, item := range items {
		if taskItem, ok := item.(ui.TableEntry); ok && !taskItem.IsProjectHeader {
			if fmt.Sprintf("%d", taskItem.ProjectID) == m.projectID &&
				fmt.Sprintf("%d", taskItem.TaskID) == m.taskID {
				m.taskList.Select(i)
				break
			}
		}
	}
}

func (m *Model) setMessage(message string, isError bool) {
	if isError {
		m.errorMsg = message
	} else {
		m.succesMsg = message
	}

	// Reset any existing timer
	if m.messageTimer != nil {
		m.messageTimer.Stop()
	}

	// Create new timer
	m.messageTimer = time.NewTimer(2 * time.Second)
}
