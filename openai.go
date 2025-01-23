package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"slices"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

var openaiClient *openai.Client
var openaiUser string

type proximityMeta struct {
	Name    string
	Score   float64
	Content string
}

func InitOpenaiClient() error {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_APIKEY")),
	)

	if client == nil {
		return fmt.Errorf("failed to create openai api client")
	}

	openaiClient = client

	openaiUser = os.Getenv("OPENAI_USER")

	return nil
}

func GenerateEmbedding(content string) ([]float64, error) {
	response, err := openaiClient.Embeddings.New(context.TODO(), openai.EmbeddingNewParams{
		Input:          openai.F[openai.EmbeddingNewParamsInputUnion](shared.UnionString(content)),
		Model:          openai.F(openai.EmbeddingModelTextEmbeddingAda002),
		EncodingFormat: openai.F(openai.EmbeddingNewParamsEncodingFormatFloat),
		User:           openai.F(openaiUser),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding from openai: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("embedding request openai returned 0 results")
	}

	return response.Data[0].Embedding, nil
}

func ByteBlobEncodeEmbedding(embedding []float64) ([]byte, error) {
	buffer := new(bytes.Buffer)
	for _, number := range embedding {
		err := binary.Write(buffer, binary.LittleEndian, number)
		if err != nil {
			return nil, fmt.Errorf("failed to littleendian binary encode float64 value: %w", err)
		}
	}

	return buffer.Bytes(), nil
}

func ByteBlobDecodeEmbedding(blob []byte) ([]float64, error) {
	var floats []float64
	buffer := bytes.NewReader(blob)
	count := buffer.Len() / 8

	for range count {
		var number float64
		err := binary.Read(buffer, binary.LittleEndian, &number)
		if err != nil {
			return nil, fmt.Errorf("failed to littleendian binary decode float64 value: %w", err)
		}
		floats = append(floats, number)
	}

	return floats, nil
}

// calculates cosine similarity (magnitude adjusted dot product) between two vectors that must be the same size
// I 100% copy/pasted this. Don't have the underlying grasp of trigonometry to write this myself
// I'll understand it at some point
func calculateProximity(a []float64, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("could not calculate proximity, inputs are not the same length")
	}

	var aMag float64
	var bMag float64
	var dotProduct float64

	for i := range len(a) {
		aMag += a[i] * a[i]
		bMag += b[i] * b[i]
		dotProduct += a[i] * b[i]
	}

	return dotProduct / (math.Sqrt(aMag) * math.Sqrt(bMag)), nil
}

func DoPrompt(prompt string) (string, error) {
	promptEmbedding, err := GenerateEmbedding(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate embedding from prompt: %w", err)
	}

	closest, err := getClosestToPrompt(promptEmbedding)
	if err != nil {
		return "", fmt.Errorf("failed to calculate the closest embedding to the prompt: %w", err)
	}

	fullPrompt := fmt.Sprintf(`
	Use the following information to answer the subsequent question.
	Information:
	%s
	
	Question:
	%s`, closest, prompt)

	chatCompletion, err := openaiClient.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fullPrompt),
		}),
		Model: openai.F(openai.ChatModelGPT3_5Turbo),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get chat completion from openai: %w", err)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

func getClosestToPrompt(promptEmbedding []float64) (string, error) {
	animalEmbeddings, err := GetAllAnimalEmbeddings()
	if err != nil {
		return "", fmt.Errorf("failed to get all animal embeddings: %w", err)
	}

	proximities := make([]proximityMeta, 0)
	for _, animalEmbedding := range animalEmbeddings {
		contentEmbedding, err := ByteBlobDecodeEmbedding(animalEmbedding.Embedding)
		if err != nil {
			return "", fmt.Errorf("failed to decode content embedding for animal %s: %w", animalEmbedding.Name, err)
		}

		score, err := calculateProximity(promptEmbedding, contentEmbedding)
		if err != nil {
			return "", fmt.Errorf("failed to calculate proximity for animal %s: %w", animalEmbedding.Name, err)
		}

		proximities = append(proximities, proximityMeta{
			Name:    animalEmbedding.Name,
			Score:   score,
			Content: animalEmbedding.Content,
		})
	}

	slices.SortFunc(proximities, func(a proximityMeta, b proximityMeta) int {
		return int(100.0 * (a.Score - b.Score))
	})

	closest := proximities[len(proximities)-1]

	return closest.Content, nil
}
