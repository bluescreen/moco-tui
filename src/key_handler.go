package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

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
		cmd := m.form.Update(msg)
		return cmd
	} else if m.focusedPane == "timeEntries" {
		m.timeEntriesTable, _ = m.timeEntriesTable.Update(msg)
		m.updateSelectedEntry()
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
		m.blurAllInputs()
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
	if m.focusedPane == "left" {
		var cmd tea.Cmd
		m.taskList, cmd = m.taskList.Update(msg)
		m.updateTaskInfo()
		return cmd
	} else if m.focusedPane == "form" {
		cmd := m.form.Update(msg)
		return cmd
	} else if m.focusedPane == "timeEntries" {
		m.timeEntriesTable, _ = m.timeEntriesTable.Update(msg)
		m.updateSelectedEntry()
	}
	return nil
}

func (m *Model) handleDownKey(msg tea.KeyMsg) tea.Cmd {
	if m.focusedPane == "left" {
		var cmd tea.Cmd
		m.taskList, cmd = m.taskList.Update(msg)
		m.updateTaskInfo()
		return cmd
	} else if m.focusedPane == "form" {
		cmd := m.form.Update(msg)
		return cmd
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
		m.blurAllInputs()
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
		if !m.confirmDelete {
			m.confirmDelete = true
			return nil
		}
		m.handleDeleteTimeEntry()
	}
	return nil
}
