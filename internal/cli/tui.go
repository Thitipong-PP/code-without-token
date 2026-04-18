package cli

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Minimalist Modern Styling ---
var (
	// Clean, modern palette
	accentColor = lipgloss.Color("#38BDF8") // Sky Blue
	textColor   = lipgloss.Color("#E4E4E7") // Zinc 200 (Soft White)
	subtleColor = lipgloss.Color("#52525B") // Zinc 600 (Muted Grey)

	// Typography styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			MarginTop(1).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			Foreground(textColor)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	checkedStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	uncheckedStyle = lipgloss.NewStyle().
			Foreground(subtleColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			MarginTop(1).
			MarginBottom(1)
)

// --- Main Menu Model ---

type model struct {
	choices  []string
	cursor   int
	selected string
}

func initialModel() model {
	return model{
		choices: []string{"Run Task", "Select Files", "Exit"},
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.choices[m.cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	// Clean header without block backgrounds
	b.WriteString(titleStyle.Render("Code Without Token"))
	b.WriteString("\n")

	for i, choice := range m.choices {
		if m.cursor == i {
			// Modern pointer
			b.WriteString(selectedItemStyle.Render(fmt.Sprintf(" ❯ %s", choice)))
		} else {
			b.WriteString(itemStyle.Render(fmt.Sprintf("   %s", choice)))
		}
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render(" ↑/↓: navigate • enter: select • q: quit"))
	return b.String()
}

func RunMainMenu() string {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}

	if finalModel, ok := m.(model); ok {
		return finalModel.selected
	}
	return ""
}

// --- File Picker Model ---

type filePickerModel struct {
	files    []string
	cursor   int
	selected map[int]struct{}
	done     bool
}

func initialFilePickerModel(files []string) filePickerModel {
	return filePickerModel{
		files:    files,
		selected: make(map[int]struct{}),
	}
}

func (m filePickerModel) Init() tea.Cmd { return nil }

func (m filePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}
		case " ":
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m filePickerModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Select Files to Include"))
	b.WriteString("\n")

	for i, file := range m.files {
		var cursor, checkbox, filename string

		// 1. Render Cursor
		if m.cursor == i {
			cursor = selectedItemStyle.Render("❯ ")
		} else {
			cursor = "  "
		}

		// 2. Render Checkbox
		if _, ok := m.selected[i]; ok {
			checkbox = checkedStyle.Render("◉ ")
		} else {
			checkbox = uncheckedStyle.Render("○ ")
		}

		// 3. Render Filename
		if m.cursor == i {
			filename = selectedItemStyle.Render(file)
		} else {
			filename = itemStyle.Render(file)
		}

		// Combine components for the row
		b.WriteString(fmt.Sprintf(" %s%s%s\n", cursor, checkbox, filename))
	}

	b.WriteString(helpStyle.Render(" space: toggle • enter: confirm • q: quit"))
	return b.String()
}

func RunFilePicker(files []string) []string {
	p := tea.NewProgram(initialFilePickerModel(files))
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error running file picker: %v\n", err)
		os.Exit(1)
	}

	if finalModel, ok := m.(filePickerModel); ok && finalModel.done {
		var selectedFiles []string
		for i, file := range finalModel.files {
			if _, ok := finalModel.selected[i]; ok {
				selectedFiles = append(selectedFiles, file)
			}
		}
		return selectedFiles
	}

	os.Exit(0)
	return nil
}