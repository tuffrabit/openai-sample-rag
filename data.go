package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

var db *sql.DB
var truncateAnimalStatement *sql.Stmt
var resetAutoIncrementAnimalStatement *sql.Stmt
var getAllAnimalStatement *sql.Stmt
var insertAnimalStatement *sql.Stmt
var truncateEmbeddingStatement *sql.Stmt
var resetAutoIncrementEmbeddingStatement *sql.Stmt
var insertEmbeddingStatement *sql.Stmt
var getAllAnimalEmbeddingStatement *sql.Stmt

type Animal struct {
	Id          int
	Name        string
	ChunkNumber int
	Content     string
}

type AnimalEmbedding struct {
	Name      string
	Content   string
	Embedding []byte
}

func OpenDb() error {
	mainDbFilename := "main.db"
	var err error
	db, err = sql.Open("sqlite", mainDbFilename)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", mainDbFilename, err)
	}

	_, err = db.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to exec journal_mode pragma: %w", err)
	}

	err = createTables()
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to create tables: %w", err)
	}

	err = prepareStatements()
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

func createTables() error {
	createAnimalQuery := `
	CREATE TABLE IF NOT EXISTS animal (
	id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	name TEXT,
	chunk_number INTEGER,
	content TEXT)`
	createAnimalStatement, err := db.Prepare(createAnimalQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare create animal statement: %w", err)
	}
	defer createAnimalStatement.Close()
	_, err = createAnimalStatement.Exec()
	if err != nil {
		return fmt.Errorf("failed to exec create animal statement: %w", err)
	}

	createEmbeddingQuery := `
	CREATE TABLE IF NOT EXISTS embedding (
	id INTEGER NOT NULL PRIMARY KEY,
	embedding blob)`
	createEmbeddingStatement, err := db.Prepare(createEmbeddingQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare create embedding statement: %w", err)
	}
	defer createEmbeddingStatement.Close()
	_, err = createEmbeddingStatement.Exec()
	if err != nil {
		return fmt.Errorf("failed to exec create embedding statement: %w", err)
	}

	return nil
}

func prepareStatements() error {
	var err error
	truncateAnimalQuery := `DELETE FROM animal`
	truncateAnimalStatement, err = db.Prepare(truncateAnimalQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare truncate animal statement: %w", err)
	}

	resetAutoIncrementAnimalQuery := `DELETE FROM SQLITE_SEQUENCE WHERE name='animal'`
	resetAutoIncrementAnimalStatement, err = db.Prepare(resetAutoIncrementAnimalQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare reset auto increment animal statement: %w", err)
	}

	getAllAnimalQuery := `SELECT * FROM animal`
	getAllAnimalStatement, err = db.Prepare(getAllAnimalQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare get all animal statement: %w", err)
	}

	insertAnimalQuery := `INSERT INTO animal(name, chunk_number, content) VALUES(?,?,?)`
	insertAnimalStatement, err = db.Prepare(insertAnimalQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare insert animal statement: %w", err)
	}

	truncateEmbeddingQuery := `DELETE FROM embedding`
	truncateEmbeddingStatement, err = db.Prepare(truncateEmbeddingQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare truncate embedding statement: %w", err)
	}

	resetAutoIncrementEmbeddingQuery := `DELETE FROM SQLITE_SEQUENCE WHERE name='embedding'`
	resetAutoIncrementEmbeddingStatement, err = db.Prepare(resetAutoIncrementEmbeddingQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare reset auto increment embedding statement: %w", err)
	}

	insertEmbeddingQuery := `INSERT INTO embedding VALUES(?,?)`
	insertEmbeddingStatement, err = db.Prepare(insertEmbeddingQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare insert embedding statement: %w", err)
	}

	getAllAnimalEmbeddingQuery := `
	SELECT animal.name, animal.content, embedding.embedding
	FROM animal
	INNER JOIN embedding
	ON animal.id = embedding.id`
	getAllAnimalEmbeddingStatement, err = db.Prepare(getAllAnimalEmbeddingQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare get all animal embedding statement: %w", err)
	}

	return nil
}

func TruncateAnimal() error {
	_, err := truncateAnimalStatement.Exec()
	if err != nil {
		return fmt.Errorf("failed to execute truncate animal statement: %w", err)
	}
	_, err = resetAutoIncrementAnimalStatement.Exec()
	if err != nil {
		return fmt.Errorf("failed to execute reset auto increment animal statement: %w", err)
	}

	return nil
}

func TruncateEmbedding() error {
	_, err := truncateEmbeddingStatement.Exec()
	if err != nil {
		return fmt.Errorf("failed to execute truncate embedding statement: %w", err)
	}
	_, err = resetAutoIncrementEmbeddingStatement.Exec()
	if err != nil {
		return fmt.Errorf("failed to execute reset auto increment embedding statement: %w", err)
	}

	return nil
}

func InsertAnimal(name string, chunkNumber int, content string) error {
	_, err := insertAnimalStatement.Exec(name, chunkNumber, content)
	if err != nil {
		return fmt.Errorf("failed to execute insert animal statement: %w", err)
	}

	return nil
}

func GetAllAnimals() ([]Animal, error) {
	rows, err := getAllAnimalStatement.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to execute get all animal query: %w", err)
	}

	animals := make([]Animal, 0)

	for rows.Next() {
		var id int
		var name string
		var chunkNumber int
		var content string

		err = rows.Scan(&id, &name, &chunkNumber, &content)
		if err != nil {
			return nil, fmt.Errorf("failed to scan animal row: %w", err)
		}

		animals = append(animals, Animal{
			Id:          id,
			Name:        name,
			ChunkNumber: chunkNumber,
			Content:     content,
		})
	}

	return animals, nil
}

func InsertEmbedding(id int, blob []byte) error {
	_, err := insertEmbeddingStatement.Exec(id, blob)
	if err != nil {
		return fmt.Errorf("failed to execute insert embedding statement: %w", err)
	}

	return nil
}

func GetAllAnimalEmbeddings() ([]AnimalEmbedding, error) {
	rows, err := getAllAnimalEmbeddingStatement.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to execute get all animal embedding query: %w", err)
	}

	animalEmbeddings := make([]AnimalEmbedding, 0)

	for rows.Next() {
		var name string
		var content string
		var embedding []byte

		err = rows.Scan(&name, &content, &embedding)
		if err != nil {
			return nil, fmt.Errorf("failed to scan animal row: %w", err)
		}

		animalEmbeddings = append(animalEmbeddings, AnimalEmbedding{
			Name:      name,
			Content:   content,
			Embedding: embedding,
		})
	}

	return animalEmbeddings, nil
}
