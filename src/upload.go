package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseCSV parses a CSV file and returns a slice of names (first column, skipping header if present)
func ParseCSV(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	var names []string
	for i, rec := range records {
		if len(rec) == 0 {
			continue
		}
		// Skip header if present
		if i == 0 && (rec[0] == "name" || rec[0] == "Name") {
			continue
		}
		names = append(names, rec[0])
	}
	return names, nil
}

// ParseJSON expects a JSON array of objects with a "name" field
func ParseJSON(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var arr []map[string]any
	if err := json.Unmarshal(data, &arr); err != nil {
		return nil, err
	}
	var names []string
	for _, obj := range arr {
		if name, ok := obj["name"].(string); ok {
			names = append(names, name)
		}
	}
	return names, nil
}

// ParseYAML expects a YAML list of objects with a "name" field
func ParseYAML(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var arr []map[string]any
	if err := yaml.Unmarshal(data, &arr); err != nil {
		return nil, err
	}
	var names []string
	for _, obj := range arr {
		if name, ok := obj["name"].(string); ok {
			names = append(names, name)
		}
	}
	return names, nil
}

// --- Exercises Bulk Upload ---
// CSV format:
// Name,Description,Category,Equipment,Types,Muscles
// Push-up,A bodyweight exercise...,Chest,Bodyweight,"Strength","Chest;Triceps"

type ExerciseUploadRow struct {
	Name        string
	Description string
	Category    string
	Equipment   []string
	Types       []string // split by ;
	Muscles     []string // split by ;
}

func ParseExercisesCSV(path string) ([]ExerciseUploadRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 1 {
		return nil, errors.New("no records found")
	}

	// Header: Name,Description,Category,Equipment,Types,Muscles
	var rows []ExerciseUploadRow
	for i, rec := range records {
		if i == 0 {
			continue // skip header
		}
		if len(rec) < 6 {
			continue
		}
		row := ExerciseUploadRow{
			Name:        strings.TrimSpace(rec[0]),
			Description: strings.TrimSpace(rec[1]),
			Category:    strings.TrimSpace(rec[2]),
			Equipment:   SplitAndTrim(rec[3], ";"), // now as []string
			Types:       SplitAndTrim(rec[4], ";"),
			Muscles:     SplitAndTrim(rec[5], ";"),
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// SplitAndTrim splits a string by sep, trims spaces and quotes
func SplitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(strings.Trim(p, `"`))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func InsertExercises(db *sql.DB, rows []ExerciseUploadRow) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	for _, row := range rows {
		// Category
		catID, err := GetOrInsertCategory(tx, row.Category)
		if err != nil {
			return fmt.Errorf("category %s: %w", row.Category, err)
		}

		// Insert exercise (no equipment_id)
		var exID int
		err = tx.QueryRow(
			`INSERT INTO exercise (name, description, category_id) 
			 VALUES ($1, $2, $3)
			 ON CONFLICT (name) DO UPDATE SET description=EXCLUDED.description 
			 RETURNING id`,
			row.Name, row.Description, catID,
		).Scan(&exID)
		if err != nil {
			return fmt.Errorf("insert exercise %s: %w", row.Name, err)
		}

		for _, e := range row.Equipment {
			e = strings.TrimSpace(e)
			if e == "" || strings.EqualFold(e, "None") {
				continue
			}
			equipID, err := GetOrInsertEquipment(tx, e)
			if err != nil {
				return fmt.Errorf("equipment %s: %w", e, err)
			}
			_, err = tx.Exec(
				`INSERT INTO exercise_equipment (exercise_id, equipment_id) 
		 VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				exID, equipID,
			)
			if err != nil {
				return fmt.Errorf("insert equipment junction: %w", err)
			}
		}

		// Types (training_type)
		for _, t := range row.Types {
			typeID, err := GetOrInsertType(tx, t)
			if err != nil {
				return fmt.Errorf("type %s: %w", t, err)
			}
			_, err = tx.Exec(
				`INSERT INTO exercise_training_types (exercise_id, training_type_id) 
				 VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				exID, typeID,
			)
			if err != nil {
				return fmt.Errorf("insert type junction: %w", err)
			}
		}

		// Muscles
		for _, m := range row.Muscles {
			muscleID, err := GetOrInsertMuscle(tx, m)
			if err != nil {
				return fmt.Errorf("muscle %s: %w", m, err)
			}
			_, err = tx.Exec(
				`INSERT INTO exercise_muscles (exercise_id, muscle_group_id) 
				 VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				exID, muscleID,
			)
			if err != nil {
				return fmt.Errorf("insert muscle junction: %w", err)
			}
		}
	}
	return nil
}

func GetOrInsertCategory(tx *sql.Tx, name string) (int, error) {
	var id int
	err := tx.QueryRow(`INSERT INTO exercise_category (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id`, name).Scan(&id)
	return id, err
}

func GetOrInsertEquipment(tx *sql.Tx, name string) (int, error) {
	var id int
	err := tx.QueryRow(`INSERT INTO equipment (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id`, name).Scan(&id)
	return id, err
}

func GetOrInsertType(tx *sql.Tx, name string) (int, error) {
	var id int
	err := tx.QueryRow(`INSERT INTO training_type (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id`, name).Scan(&id)
	return id, err
}

func GetOrInsertMuscle(tx *sql.Tx, name string) (int, error) {
	var id int
	err := tx.QueryRow(`INSERT INTO muscle_group (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id`, name).Scan(&id)
	return id, err
}

// DB is an interface for *sql.DB or a transaction, for testability
// You can use *sql.DB directly

type DB interface {
	Exec(query string, args ...any) (any, error)
}
