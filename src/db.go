package main

import (
	"database/sql"
	"fmt"
)

const (
	InsertMuscleGroupQuery  = "INSERT INTO muscle_group (name) VALUES ($1) ON CONFLICT (name) DO NOTHING"
	InsertTrainingTypeQuery = "INSERT INTO training_type (name) VALUES ($1) ON CONFLICT (name) DO NOTHING"
	InsertCategoryQuery     = "INSERT INTO exercise_category (name) VALUES ($1) ON CONFLICT (name) DO NOTHING"
	InsertEquipmentQuery    = "INSERT INTO equipment (name) VALUES ($1) ON CONFLICT (name) DO NOTHING"
)

func InsertNamesToDB(db *sql.DB, query string, names []string) error {
	for _, name := range names {
		_, err := db.Exec(query, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetTableCount(db *sql.DB, table string) int {
	var count int
	_ = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	return count
}
