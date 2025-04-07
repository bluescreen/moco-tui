package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/denwerk/moco/src/types"
)

type Item struct {
	Title           string
	Desc            string
	TaskID          int
	ProjectID       int
	IsProjectHeader bool
	Position        int
}

func (i Item) FilterValue() string { return "" }

func MapProjectsToItems(projects []types.Project) []list.Item {
	var items []list.Item

	for _, project := range projects {
		i := 1
		if len(project.Tasks) > 0 {
			items = append(items, Item{
				Title:           project.Name,
				Desc:            "Project",
				IsProjectHeader: true,
			})

			for _, task := range project.Tasks {
				items = append(items, Item{
					Desc:            task.Name,
					Title:           project.Customer.Name,
					TaskID:          task.ID,
					ProjectID:       project.ID,
					IsProjectHeader: false,
					Position:        i,
				})
				i++
			}
		}
	}

	return items
}
