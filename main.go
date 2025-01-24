package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	initEverything()
	defer db.Close()
	populateDb := flag.Bool("populatedb", false, "populate the db with animal chunks from markdown files in the adjacent animals directory")
	prompt := flag.String("prompt", "", "your question for the AI overlord")
	skipRag := flag.Bool("skiprag", false, "skip prompt engineering with RAG")
	flag.Parse()

	if *populateDb {
		fmt.Println("Populating DB with animals and their embeddings...")
		err := populateDbWithAnimals()
		if err != nil {
			log.Panicln("failed to populate animal table:", err)
		}

		err = populateDbWithEmbeddings()
		if err != nil {
			log.Panicln("failed to populate embedding table:", err)
		}
	}

	if prompt == nil || *prompt == "" {
		fmt.Println("you have no question? ok bye")
		os.Exit(0)
	}

	response, err := DoPrompt(*prompt, *skipRag)
	if err != nil {
		log.Panicln("failed to prompt openai:", err)
	}

	fmt.Println("the AI says...")
	fmt.Println(response)
	os.Exit(0)
}

func initEverything() {
	err := godotenv.Load()
	if err != nil {
		log.Panicln("failed to load .env:", err)
	}

	err = OpenDb()
	if err != nil {
		log.Panicln("failed to open db:", err)
	}

	err = InitOpenaiClient()
	if err != nil {
		log.Panicln("failed to init openai api client:", err)
	}
}
