package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
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
	ticker           *time.Ticker
	focusedPane      string           // "left", "form", or "timeEntries"
	confirmDelete    bool             // Whether to show delete confirmation
	selectedEntry    *types.TimeEntry // Currently selected time entry
	lastUpdate       time.Time        // When time entries were last updated
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
		m.errorMsg = "Please select a project first"
		return
	}

	// Validate hours
	hours, err := parseHours(m.hours)
	if err != nil {
		m.errorMsg = err.Error()
		return
	}

	// Validate date
	if m.date == "" {
		m.errorMsg = "Date is required"
		return
	}

	// Validate description
	if m.description == "" {
		m.errorMsg = "Description is required"
		return
	}

	projectID, err := strconv.Atoi(m.projectID)
	if err != nil {
		m.errorMsg = "Invalid project ID"
		return
	}

	taskID, err := strconv.Atoi(m.taskID)
	if err != nil {
		m.errorMsg = "Invalid task ID"
		return
	}

	entry := types.TimeEntry{
		Date:        m.date,
		Hours:       hours,
		ProjectID:   projectID,
		TaskID:      taskID,
		Description: m.description,
	}

	err = submitTimeEntry(m.cfg, entry)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Error submitting time entry: %v", err)
	} else {
		m.succesMsg = "Time entry submitted successfully!"
		// Clear form
		m.date = time.Now().Format("2006-01-02")
		m.hours = "0"
		m.description = ""
		m.dateInput.SetValue(m.date)
		m.hoursInput.SetValue(m.hours)
		m.descInput.SetValue(m.description)
		// Reload time entries
		m.loadTimeEntries()
	}
}

func (m *Model) handleDeleteTimeEntry() {
	if m.selectedEntry == nil {
		return
	}

	if !m.confirmDelete {
		m.confirmDelete = true
		return
	}

	err := deleteTimeEntry(m.cfg, m.selectedEntry.ID)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Error deleting time entry: %v", err)
	} else {
		m.succesMsg = "Time entry deleted successfully!"
		m.loadTimeEntries()
	}
	m.confirmDelete = false
	m.selectedEntry = nil
}

func (m *Model) Init() tea.Cmd {
	// Start the ticker for polling time entries
	m.ticker = time.NewTicker(10 * time.Second)
	m.focusedPane = "form" // Start with focus on form pane
	m.focusedInput = 0
	m.focusCurrentInput()

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

	if m.focusedPane == "left" {
		m.taskList, cmd = m.taskList.Update(msg)
		m.updateTaskInfo()
	}

	return m, cmd
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		return m.handleEscKey()
	case "enter":
		return m.handleEnterKey()
	case "right":
		return m.handleRightKey()
	case "left":
		return m.handleLeftKey()
	case "up", "k":
		return m.handleUpKey(msg)
	case "down", "j":
		return m.handleDownKey(msg)
	case "tab":
		return m.handleTabKey()
	case "d":
		return m.handleDKey()
	}

	if m.focusedPane == "form" {
		m.updateFormInputs(msg)
	} else if m.focusedPane == "timeEntries" {
		m.timeEntriesTable, _ = m.timeEntriesTable.Update(msg)
	}

	return nil
}

func (m *Model) handleEscKey() tea.Cmd {
	if m.confirmDelete {
		m.confirmDelete = false
		m.selectedEntry = nil
		return nil
	}
	if m.focusedPane != "left" {
		m.focusedPane = "left"
		m.blurAllInputs()
		return nil
	}
	if m.ticker != nil {
		m.ticker.Stop()
	}
	return tea.Quit
}

func (m *Model) handleEnterKey() tea.Cmd {
	if m.focusedPane == "form" {
		m.handleTimeEntrySubmission()
	} else if m.confirmDelete {
		m.handleDeleteTimeEntry()
	}
	return nil
}

func (m *Model) handleRightKey() tea.Cmd {
	if m.focusedPane == "left" {
		m.focusedPane = "form"
		m.focusedInput = 0
		m.focusCurrentInput()
		m.saveLastTask()
	}
	return nil
}

func (m *Model) handleLeftKey() tea.Cmd {
	if m.focusedPane != "left" {
		m.focusedPane = "left"
	}
	return nil
}

func (m *Model) handleUpKey(msg tea.KeyMsg) tea.Cmd {
	if m.focusedPane == "form" {
		m.focusedInput = (m.focusedInput - 1 + 3) % 3
		m.focusCurrentInput()
	} else if m.focusedPane == "timeEntries" {
		m.timeEntriesTable, _ = m.timeEntriesTable.Update(msg)
		m.updateSelectedEntry()
	}
	return nil
}

func (m *Model) handleDownKey(msg tea.KeyMsg) tea.Cmd {
	if m.focusedPane == "form" {
		m.focusedInput = (m.focusedInput + 1) % 3
		m.focusCurrentInput()
	} else if m.focusedPane == "timeEntries" {
		m.timeEntriesTable, _ = m.timeEntriesTable.Update(msg)
		m.updateSelectedEntry()
	}
	return nil
}

func (m *Model) handleTabKey() tea.Cmd {
	switch m.focusedPane {
	case "left":
		m.focusedPane = "form"
		m.focusedInput = 0
		m.focusCurrentInput()
	case "form":
		m.focusedPane = "timeEntries"
		m.blurAllInputs()
	case "timeEntries":
		m.focusedPane = "left"
	}
	return nil
}

func (m *Model) handleDKey() tea.Cmd {
	if m.focusedPane == "timeEntries" && m.selectedEntry != nil {
		m.handleDeleteTimeEntry()
	}
	return nil
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
		m.focusedInput = 0
		m.focusCurrentInput()
	} else if msg.X >= leftWidth && msg.Y >= formHeight {
		m.focusedPane = "timeEntries"
		m.blurAllInputs()
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

func (m *Model) updateFormInputs(msg tea.KeyMsg) {
	m.dateInput, _ = m.dateInput.Update(msg)
	m.hoursInput, _ = m.hoursInput.Update(msg)
	m.descInput, _ = m.descInput.Update(msg)
}

func (m *Model) updateTaskInfo() {
	selected := m.taskList.SelectedItem()
	if selected == nil {
		return
	}

	selectedItem, ok := selected.(ui.Item)
	if !ok || selectedItem.IsProjectHeader {
		return
	}

	m.taskID = fmt.Sprintf("%d", selectedItem.TaskID)
	m.projectID = fmt.Sprintf("%d", selectedItem.ProjectID)
	m.taskTitle = selectedItem.Desc
}

func (m *Model) updateSelectedEntry() {
	if row := m.timeEntriesTable.SelectedRow(); len(row) > 0 {
		for _, entry := range m.timeEntries {
			if fmt.Sprintf("%d", entry.ID) == row[0] {
				m.selectedEntry = &entry
				break
			}
		}
	}
}

func (m *Model) blurAllInputs() {
	m.dateInput.Blur()
	m.hoursInput.Blur()
	m.descInput.Blur()
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
	entryForm := lipgloss.JoinVertical(lipgloss.Left,
		ui.TitleStyle.Render("New Time Entry"),
		fmt.Sprintf("Task: %s", m.taskTitle),
		fmt.Sprintf("Date: %s", m.dateInput.View()),
		fmt.Sprintf("Hours: %s", m.hoursInput.View()),
		fmt.Sprintf("Description: %s", m.descInput.View()),
		"Press 'enter' to submit, 'esc' to cancel, 'tab' to switch between panes.",
	)

	// Add error message if present
	var errorMsg string
	if m.errorMsg != "" {
		errorMsg = ui.ErrorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
	}

	// Add success message if present
	var successMsg string
	if m.succesMsg != "" {
		successMsg = ui.SuccessStyle.Render(m.succesMsg)
	}

	// Combine messages
	messages := lipgloss.JoinVertical(lipgloss.Left, errorMsg, successMsg)

	formSection := lipgloss.JoinVertical(lipgloss.Left,
		entryForm,
		messages,
	)

	formStyle := ui.PaneStyle.Width(rightWidth)
	if m.focusedPane == "form" {
		formStyle = ui.FocusedPaneStyle.Width(rightWidth)
	}

	formPane := formStyle.Render(formSection)

	// Time Entries Section
	timeEntriesTitle := ui.TitleStyle.Render("Time Entries")
	lastUpdate := ui.LastUpdateStyle.Render(fmt.Sprintf("Last updated: %s", m.lastUpdate.Format("15:04:05")))

	timeEntriesContent := lipgloss.JoinVertical(lipgloss.Left,
		timeEntriesTitle,
		lastUpdate,
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

	dateInput := textinput.New()
	dateInput.Placeholder = "Enter date (YYYY-MM-DD)"
	dateInput.SetValue(time.Now().Format("2006-01-02"))

	hoursInput := textinput.New()
	hoursInput.Placeholder = "Enter hours (e.g. 1.5 or 1:30)"

	descInput := textinput.New()
	descInput.Placeholder = "Enter description"

	hoursInput.Focus()

	model := &Model{
		cfg:        cfg,
		taskList:   taskList,
		dateInput:  dateInput,
		hoursInput: hoursInput,
		descInput:  descInput,
		date:       time.Now().Format("2006-01-02"),
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
		if taskItem, ok := item.(ui.Item); ok && !taskItem.IsProjectHeader {
			if fmt.Sprintf("%d", taskItem.ProjectID) == m.projectID &&
				fmt.Sprintf("%d", taskItem.TaskID) == m.taskID {
				m.taskList.Select(i)
				break
			}
		}
	}
}

func (m *Model) focusCurrentInput() {
	// Blur all inputs first
	m.dateInput.Blur()
	m.hoursInput.Blur()
	m.descInput.Blur()

	// Focus the current input
	switch m.focusedInput {
	case 0:
		m.dateInput.Focus()
	case 1:
		m.hoursInput.Focus()
	case 2:
		m.descInput.Focus()
	}
}
