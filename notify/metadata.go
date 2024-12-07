package notify

import (
	"github.com/apfelfrisch/zh-notify/db"
	openai "github.com/sashabaranov/go-openai"
	"golang.org/x/oauth2/clientcredentials"
)

type MetaDataService interface {
	Init() error
	SetCategory(event *db.Event) error
	SetArtist(event *db.Event) error
	SetArtistUrl(event *db.Event) error
	SetArtistImgUrl(event *db.Event) error
}

func NewMetaDataService(openAiToken string, spotifyCred clientcredentials.Config) MetaDataService {
	return &metaDataService{
		openAi: openAiParser{
			client:    openai.NewClient(openAiToken),
			initPromt: nil,
			response:  nil,
		},
		spotify: spotifyService{
			&spotifyCred,
			nil,
			nil,
		},
	}
}

type metaDataService struct {
	openAi  openAiParser
	spotify spotifyService
}

func (md *metaDataService) Init() error {
	if err := md.openAi.Init(); err != nil {
		return err
	}
	if err := md.spotify.Init(); err != nil {
		return err
	}
	return nil
}

func (md *metaDataService) SetArtist(event *db.Event) error {
	return md.openAi.SetArtist(event)
}

func (md *metaDataService) SetCategory(event *db.Event) error {
	return md.openAi.SetCategory(event)
}

func (md *metaDataService) SetArtistUrl(event *db.Event) error {
	return md.spotify.SetArtistUrl(event)
}

func (md *metaDataService) SetArtistImgUrl(event *db.Event) error {
	return md.spotify.SetArtistImgUrl(event)
}
