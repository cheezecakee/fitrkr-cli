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
	db           *sql.DB
	counts       []int
}

var menuOptions = []string{
	"Upload Muscle Groups",
	"Upload Exercise Types",
	"Upload Exercise Categories",
	"Upload Equipments",
	"Upload Exercises",
	"Quit",
}

func initialModel(db *sql.DB) model {
	return model{
		state:      stateMenu,
		menuChoice: 0,
		db:         db,
	}
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
					return m, nil
				}
				m.state = stateFileSelector
				m.fileList = append(files, "Back")
				m.fileChoice = 0
				return m, nil
			}
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
			// Detect file type
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
				return m, nil
			}

			// Pick the right query
			var query string
			switch m.menuChoice {
			case 0:
				query = InsertMuscleGroupQuery
			case 1:
				query = InsertExerciseTypeQuery
			case 2:
				query = InsertCategoryQuery
			case 3:
				query = InsertEquipmentQuery
			case 4: // assuming "Upload Exercises" is the 5th option (index 4)
				rows, err := ParseExercisesCSV(m.selectedFile)
				if err != nil {
					m.state = stateResult
					m.resultMsg = fmt.Sprintf("Error parsing exercises CSV: %v\nPress enter or q to return to menu.", err)
					return m, nil
				}
				err = InsertExercises(m.db, rows)
				if err != nil {
					m.state = stateResult
					m.resultMsg = fmt.Sprintf("DB error: %v\nPress enter or q to return to menu.", err)
					return m, nil
				}
				m.state = stateResult
				m.resultMsg = fmt.Sprintf("Successfully uploaded %d exercises!\nPress enter or q to return to menu.", len(rows))
				return m, nil
			}

			err = InsertNamesToDB(m.db, query, names)
			if err != nil {
				m.state = stateResult
				m.resultMsg = fmt.Sprintf("DB error: %v\nPress enter or q to return to menu.", err)
				return m, nil
			}

			m.state = stateResult
			m.resultMsg = fmt.Sprintf("Successfully uploaded %d entries!\nPress enter or q to return to menu.", len(names))
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateMenu:
		var b strings.Builder
		b.WriteString("\n  Select:\n\n")
		for i, opt := range menuOptions {
			cursor := " "
			if i == m.menuChoice {
				cursor = ">"
			}
			countStr := ""
			if i < len(m.counts) {
				countStr = fmt.Sprintf(" [%3d]", m.counts[i])
			}
			b.WriteString(fmt.Sprintf("  %s %-30s%s\n", cursor, opt, countStr))
		}
		b.WriteString("\n  Use up/down (j/k) to move, enter to select.\n")
		return b.String()
	case stateFileSelector:
		var b strings.Builder
		b.WriteString("\n  Pick a file to upload:\n\n")
		for i, fname := range m.fileList {
			cursor := " "
			if i == m.fileChoice {
				cursor = ">"
			}
			b.WriteString(fmt.Sprintf("  %s %s\n", cursor, fname))
		}
		b.WriteString("\n  Use up/down (j/k) to move, enter to select, q/esc to go back.\n")
		return b.String()
	case stateResult:
		return "\n  " + m.resultMsg + "\n"
	default:
		return ""
	}
}

// listDataFiles returns a sorted list of .csv, .json, .yaml, .yml files in ./data/
func listDataFiles() ([]string, error) {
	entries, err := os.ReadDir("./src/internal/data/")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".csv" || ext == ".json" || ext == ".yaml" || ext == ".yml" {
			files = append(files, name)
		}
	}
	sort.Strings(files)
	return files, nil
}

func InitMenu(db *sql.DB) {
	p := tea.NewProgram(initialModel(db))
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}

func (m *model) RefreshCounts() {
	m.counts = []int{
		GetTableCount(m.db, "muscle_groups"),
		GetTableCount(m.db, "exercise_types"),
		GetTableCount(m.db, "exercise_categories"),
		GetTableCount(m.db, "equipment"),
		GetTableCount(m.db, "exercises"),
	}
}
