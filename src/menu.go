package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type appState int

const (
	stateMenu appState = iota
	stateFileSelector
	stateResult
)

type model struct {
	state        appState
	menuChoice   int
	fileList     []string
	fileChoice   int
	selectedFile string
	resultMsg    string
	isError      bool
	db           *sql.DB
	counts       []int
}

var menuOptions = []string{
	"Upload Muscle Groups",
	"Upload Exercise Types",
	"Upload Exercise Categories",
	"Upload Equipment",
	"Upload Exercises",
	"Quit",
}

func initialModel(db *sql.DB) model {
	m := model{
		state:      stateMenu,
		menuChoice: 0,
		db:         db,
	}
	// Initialize counts on startup
	m.refreshCounts()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateMenu:
		return updateMenu(m, msg)
	case stateFileSelector:
		return updateFileMenu(m, msg)
	case stateResult:
		if key, ok := msg.(tea.KeyMsg); ok && (key.String() == "enter" || key.String() == "q" || key.String() == "esc") {
			m.state = stateMenu
			m.resultMsg = ""
			m.isError = false
			// Refresh counts when returning to menu
			m.refreshCounts()
			return m, nil
		}
		return m, nil
	default:
		return m, nil
	}
}

func updateMenu(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.menuChoice > 0 {
				m.menuChoice--
			}
		case "down", "j":
			if m.menuChoice < len(menuOptions)-1 {
				m.menuChoice++
			}
		case "enter":
			if m.menuChoice == len(menuOptions)-1 {
				return m, tea.Quit
			} else {
				// List files in ./src/internal/data/
				files, err := listDataFiles()
				if err != nil {
					m.state = stateResult
					m.resultMsg = fmt.Sprintf("Error reading ./src/internal/data/: %v\nPress enter or q to return to menu.", err)
					m.isError = true
					return m, nil
				}
				m.state = stateFileSelector
				m.fileList = append(files, "Back")
				m.fileChoice = 0
				return m, nil
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.refreshCounts()
			return m, nil
		}
	}
	return m, nil
}

func updateFileMenu(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.fileChoice > 0 {
				m.fileChoice--
			}
		case "down", "j":
			if m.fileChoice < len(m.fileList)-1 {
				m.fileChoice++
			}
		case "q", "esc":
			m.state = stateMenu
			return m, nil
		case "enter":
			if m.fileList[m.fileChoice] == "Back" {
				m.state = stateMenu
				return m, nil
			}
			m.selectedFile = filepath.Join("./src/internal/data/", m.fileList[m.fileChoice])

			// Detect file type and process
			ext := strings.ToLower(filepath.Ext(m.selectedFile))
			var names []string
			var err error

			switch ext {
			case ".csv":
				names, err = ParseCSV(m.selectedFile)
			case ".json":
				names, err = ParseJSON(m.selectedFile)
			case ".yaml", ".yml":
				names, err = ParseYAML(m.selectedFile)
			default:
				err = fmt.Errorf("unsupported file type: %s", ext)
			}

			if err != nil {
				m.state = stateResult
				m.resultMsg = fmt.Sprintf("Error parsing file: %v\nPress enter or q to return to menu.", err)
				m.isError = true
				return m, nil
			}

			// Handle different upload types
			var query string
			switch m.menuChoice {
			case 0: // Muscle Groups
				query = InsertMuscleGroupQuery
			case 1: // Exercise Types
				query = InsertTrainingTypeQuery
			case 2: // Exercise Categories
				query = InsertCategoryQuery
			case 3: // Equipment
				query = InsertEquipmentQuery
			case 4: // Exercises (special handling)
				rows, err := ParseExercisesCSV(m.selectedFile)
				if err != nil {
					m.state = stateResult
					m.resultMsg = fmt.Sprintf("Error parsing exercises CSV: %v\nPress enter or q to return to menu.", err)
					m.isError = true
					return m, nil
				}
				err = InsertExercises(m.db, rows)
				if err != nil {
					m.state = stateResult
					m.resultMsg = fmt.Sprintf("Database error: %v\nPress enter or q to return to menu.", err)
					m.isError = true
					return m, nil
				}
				m.state = stateResult
				m.resultMsg = fmt.Sprintf("Successfully uploaded %d exercises!\nPress enter or q to return to menu.", len(rows))
				m.isError = false
				return m, nil
			}

			// Insert simple name-based entries
			err = InsertNamesToDB(m.db, query, names)
			if err != nil {
				m.state = stateResult
				m.resultMsg = fmt.Sprintf("Database error: %v\nPress enter or q to return to menu.", err)
				m.isError = true
				return m, nil
			}

			m.state = stateResult
			m.resultMsg = fmt.Sprintf("Successfully uploaded %d entries!\nPress enter or q to return to menu.", len(names))
			m.isError = false
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateMenu:
		var parts []string

		// App header
		parts = append(parts, RenderAppHeader())
		parts = append(parts, "")

		// Menu title
		parts = append(parts, RenderMenuTitle("Select an option:"))
		parts = append(parts, "")

		// Menu items
		for i, opt := range menuOptions {
			count := -1
			if i < len(m.counts) {
				count = m.counts[i]
			}
			parts = append(parts, RenderMenuItem(opt, i == m.menuChoice, count))
		}

		// Help text
		parts = append(parts, "")
		parts = append(parts, RenderHelpText("Navigation: ↑/↓ or j/k • Select: enter • Refresh counts: r • Quit: q"))

		return ContainerStyle.Render(strings.Join(parts, "\n"))

	case stateFileSelector:
		var parts []string

		// File selector title
		parts = append(parts, RenderMenuTitle("Select a file to upload:"))
		parts = append(parts, "")

		// File list
		for i, filename := range m.fileList {
			isBackOption := filename == "Back"
			parts = append(parts, RenderFileItem(filename, i == m.fileChoice, isBackOption))
		}

		// Help text
		parts = append(parts, "")
		parts = append(parts, RenderHelpText("Navigation: ↑/↓ or j/k • Select: enter • Back: q/esc"))

		return ContainerStyle.Render(strings.Join(parts, "\n"))

	case stateResult:
		var content string
		if m.isError {
			content = RenderErrorMessage(m.resultMsg)
		} else {
			content = RenderSuccessMessage(m.resultMsg)
		}

		// Add help text
		content += "\n\n" + RenderHelpText("Press enter, q, or esc to continue")

		return ContainerStyle.Render(content)

	default:
		return ""
	}
}

// listDataFiles returns a sorted list of supported files in ./data/
func listDataFiles() ([]string, error) {
	entries, err := os.ReadDir("./src/internal/data/")
	if err != nil {
		return nil, err
	}

	var files []string
	supportedExts := map[string]bool{
		".csv":  true,
		".json": true,
		".yaml": true,
		".yml":  true,
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))

		if supportedExts[ext] {
			files = append(files, name)
		}
	}

	sort.Strings(files)
	return files, nil
}

func InitMenu(db *sql.DB) {
	p := tea.NewProgram(initialModel(db))
	if _, err := p.Run(); err != nil {
		fmt.Println("Error starting program:", err)
		os.Exit(1)
	}
}

// refreshCounts updates the database table counts for display
func (m *model) refreshCounts() {
	m.counts = []int{
		GetTableCount(m.db, "muscle_group"),
		GetTableCount(m.db, "training_type"),
		GetTableCount(m.db, "exercise_category"),
		GetTableCount(m.db, "equipment"),
		GetTableCount(m.db, "exercise"),
	}
}
