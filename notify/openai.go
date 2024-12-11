package notify

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/apfelfrisch/zh-notify/db"
	openai "github.com/sashabaranov/go-openai"
)

const INIT_PROMT string = `Filter aus der Ankündigung den "Interpreten" und die "Kategorie" der Veranstaltung.
- Folgende Kategorien stehen zur Verfügung: concert, reading, theatre, comedy, party, unkown.
- Der Text zwischen () muss ignoriert werden.
- Ignoriere "& Band".
- Umschließe die Antwort nicht mit JSON-Markierungen.
Antworte im folgendem json format: {"artist": "Interpreten", "category": "Kategorie"}`

type metaData struct {
	Artist   string `json:"artist"`
	Category string `json:"category"`
}

type openaiResp struct {
	event    db.Event
	metaData metaData
}

type openAiParser struct {
	client    *openai.Client
	initPromt *openai.ChatCompletionMessage
	response  *openaiResp
}

func (oai *openAiParser) Init() error {
	message := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: INIT_PROMT,
	}

	_, err := oai.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{message},
		},
	)

	oai.initPromt = &message

	return err
}

func (oai *openAiParser) setArtist(event *db.Event) error {
	if event.Artist.Valid {
		return nil
	}

	metaData, err := oai.requestHeadlineParsing(event)

	if err != nil {
		return err
	}

	if metaData.Artist != "" {
		event.Artist = sql.NullString{String: metaData.Artist, Valid: true}
	}

	return nil
}

func (oai *openAiParser) setCategory(event *db.Event) error {
	if event.Category.Valid {
		return nil
	}

	metaData, err := oai.requestHeadlineParsing(event)

	if err != nil {
		return err
	}

	if metaData.Category != "" {
		event.Category = sql.NullString{String: metaData.Category, Valid: true}
	}

	return nil
}

func (oai *openAiParser) requestHeadlineParsing(event *db.Event) (metaData, error) {
	if oai.response != nil && oai.response.event.ID == event.ID {
		return oai.response.metaData, nil
	}

	messages := []openai.ChatCompletionMessage{
		*oai.initPromt,
		{
			Role:    openai.ChatMessageRoleUser,
			Content: event.Name,
		},
	}

	resp, err := oai.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)

	var md metaData

	if err != nil || len(resp.Choices) != 1 {
		return md, err
	}

	json.Unmarshal([]byte(resp.Choices[0].Message.Content), &md)

	oai.response = &openaiResp{
		event:    *event,
		metaData: md,
	}

	return md, nil
}
