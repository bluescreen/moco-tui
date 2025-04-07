package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles used throughout the application
var (
	// DocStyle is used for the main layout
	DocStyle = lipgloss.NewStyle().Margin(1, 2)

	// TitleStyle is used for headers
	TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))

	HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))

	TotalStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))

	// ItemStyle is used for list items
	ItemStyle = lipgloss.NewStyle().PaddingLeft(4)

	// SelectedItemStyle is used for selected list items
	SelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("255"))

	// ErrorStyle is used for error messages
	ErrorStyle = lipgloss.NewStyle().PaddingTop(1).Foreground(lipgloss.Color("196"))

	// SuccessStyle is used for success messages
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))

	// LastUpdateStyle is used for the last update timestamp
	LastUpdateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// PaneStyle is used for panes
	PaneStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 0)

	// FocusedPaneStyle is used for focused panes
	FocusedPaneStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("255")).
				Padding(0, 0)

	// ConfirmDialogStyle is used for confirmation dialogs
	ConfirmDialogStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("196")). // Red border for warning
				Padding(1, 0)
)

type ItemDelegate struct{}

func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}
	str := ""
	fn := ItemStyle.Render

	if i.IsProjectHeader {
		str = fmt.Sprintf("%s", i.Title)
	} else {
		str = fmt.Sprintf("\t [%d] %s", i.Position, i.Desc)

		if index == m.Index() {
			fn = func(s ...string) string {
				return SelectedItemStyle.Render("> " + strings.Join(s, " "))
			}
		}
	}

	fmt.Fprint(w, fn(str))
}
