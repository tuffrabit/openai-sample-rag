package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

func populateDbWithAnimals() error {
	encoder, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return fmt.Errorf("could not get tiktoken encoder: %w", err)
	}

	animalsDirPath := "./animals"
	_, err = os.Stat(animalsDirPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("animals directory must exist: %w", err)
	}

	entries, err := os.ReadDir(animalsDirPath)
	if err != nil {
		return fmt.Errorf("could not read animals directory: %w", err)
	}

	err = TruncateAnimal()
	if err != nil {
		return fmt.Errorf("failed to truncate animal table: %w", err)
	}
	err = TruncateEmbedding()
	if err != nil {
		return fmt.Errorf("failed to truncate embedding table: %w", err)
	}

	for i, entry := range entries {
		if entry.IsDir() {
			continue
		}

		entryName := entry.Name()
		if !strings.EqualFold(filepath.Ext(entryName), ".md") {
			continue
		}

		animalName := strings.TrimSuffix(entryName, ".md")
		fullFilePath := animalsDirPath + "/" + entryName
		animalContent, err := os.ReadFile(fullFilePath)
		if err != nil {
			return fmt.Errorf("could not read contents of %s: %w", fullFilePath, err)
		}

		tokens := encoder.Encode(string(animalContent), nil, nil)
		if len(tokens) > 1000 {
			return fmt.Errorf("chunk token size of %d to large for %s", len(tokens), entryName)
		}

		err = InsertAnimal(animalName, i, string(animalContent))
		if err != nil {
			return fmt.Errorf("failed to insert animal %s: %w", animalName, err)
		}
	}

	return nil
}

func populateDbWithEmbeddings() error {
	animals, err := GetAllAnimals()
	if err != nil {
		return fmt.Errorf("failed to get all animals: %w", err)
	}

	for _, animal := range animals {
		embedding, err := GenerateEmbedding(animal.Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding for animal %s: %w", animal.Name, err)
		}

		encodedEmbedding, err := ByteBlobEncodeEmbedding(embedding)
		if err != nil {
			return fmt.Errorf("failed to encode embedding for animal %s: %w", animal.Name, err)
		}

		err = InsertEmbedding(animal.Id, encodedEmbedding)
		if err != nil {
			return fmt.Errorf("failed to insert embedding for animal %s: %w", animal.Name, err)
		}
	}

	return nil
}
