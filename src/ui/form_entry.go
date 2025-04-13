package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FormEntry struct {
	dateInput    textinput.Model
	hoursInput   textinput.Model
	descInput    textinput.Model
	focusedInput int
	width        int
	height       int
	taskTitle    string
}

func NewFormEntry() FormEntry {
	dateInput := textinput.New()
	dateInput.Placeholder = "Enter date (YYYY-MM-DD)"
	dateInput.SetValue(time.Now().Format("2006-01-02"))

	hoursInput := textinput.New()
	hoursInput.Placeholder = "Enter hours (e.g. 1.5 or 1:30)"

	descInput := textinput.New()
	descInput.Placeholder = "Enter description"

	return FormEntry{
		dateInput:    dateInput,
		hoursInput:   hoursInput,
		descInput:    descInput,
		focusedInput: 0,
	}
}

func (f *FormEntry) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			f.focusedInput = (f.focusedInput - 1 + 3) % 3
			f.focusCurrentInput()
		case "down":
			f.focusedInput = (f.focusedInput + 1) % 3
			f.focusCurrentInput()
		default:
			f.updateInputs(msg)
		}
	}

	return cmd
}

func (f *FormEntry) View() string {
	form := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("New Time Entry"),
		fmt.Sprintf("Task: %s", f.taskTitle),
		fmt.Sprintf("Date: %s", f.dateInput.View()),
		fmt.Sprintf("Hours: %s", f.hoursInput.View()),
		fmt.Sprintf("Description: %s", f.descInput.View()),
		"Press 'enter' to submit, 'esc' to cancel, 'tab' to switch between panes.",
	)

	return form
}

func (f *FormEntry) focusCurrentInput() {
	// Blur all inputs first
	f.dateInput.Blur()
	f.hoursInput.Blur()
	f.descInput.Blur()

	// Focus the current input
	switch f.focusedInput {
	case 0:
		f.dateInput.Focus()
	case 1:
		f.hoursInput.Focus()
	case 2:
		f.descInput.Focus()
	}
}

func (f *FormEntry) updateInputs(msg tea.KeyMsg) {
	f.dateInput, _ = f.dateInput.Update(msg)
	f.hoursInput, _ = f.hoursInput.Update(msg)
	f.descInput, _ = f.descInput.Update(msg)
}

func (f *FormEntry) BlurAll() {
	f.dateInput.Blur()
	f.hoursInput.Blur()
	f.descInput.Blur()
}

func (f *FormEntry) Clear() {
	f.dateInput.SetValue(time.Now().Format("2006-01-02"))
	f.hoursInput.SetValue("")
	f.descInput.SetValue("")
}

func (f *FormEntry) GetValues() (string, string, string) {
	return f.dateInput.Value(), f.hoursInput.Value(), f.descInput.Value()
}

func (f *FormEntry) SetSize(width, height int) {
	f.width = width
	f.height = height
}

func (f *FormEntry) SetTaskTitle(title string) {
	f.taskTitle = title
}
