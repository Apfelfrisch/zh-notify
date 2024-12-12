package internal

import (
	"github.com/apfelfrisch/zh-notify/internal/collect"
	"github.com/apfelfrisch/zh-notify/internal/collect/openai"
	"github.com/apfelfrisch/zh-notify/internal/collect/spotify"
	"github.com/apfelfrisch/zh-notify/internal/db"
)

const URL = "https://www.zollhaus-leer.com/veranstaltungen/"

func CollectNewEvents() ([]collect.Event, error) {
	return collect.CrawlEvents(URL)
}

func NewSyncEventCollector(openAiToken, spotifyId, sporitySecret string) *SyncCollector {
	return &SyncCollector{
		service: &syncService{
			OpenAi:  openai.New(openAiToken),
			Spotify: spotify.New(spotifyId, sporitySecret),
		},
	}
}

type SyncCollector struct {
	service collect.EventSyncCollector
}

func (sc *SyncCollector) Init() error {
	return sc.service.Init()
}

func (sc *SyncCollector) Sync(event *db.Event) error {
	if !event.Artist.Valid {
		if err := sc.service.SetArtist(event); err != nil {
			return err
		}
	}

	if !event.Category.Valid {
		if err := sc.service.SetCategory(event); err != nil {
			return err
		}
	}

	if !event.ArtistUrl.Valid {
		if err := sc.service.SetArtistUrl(event); err != nil {
			return err
		}
	}

	if !event.ArtistImgUrl.Valid {
		if err := sc.service.SetArtistImgUrl(event); err != nil {
			return err
		}
	}

	return nil
}

type syncService struct {
	OpenAi  *openai.Service
	Spotify *spotify.Service
}

func (md *syncService) Init() error {
	if err := md.OpenAi.Init(); err != nil {
		return err
	}
	if err := md.Spotify.Init(); err != nil {
		return err
	}
	return nil
}

func (md *syncService) SetArtist(event *db.Event) error {
	return md.OpenAi.SetArtist(event)
}

func (md *syncService) SetCategory(event *db.Event) error {
	return md.OpenAi.SetCategory(event)
}

func (md *syncService) SetArtistUrl(event *db.Event) error {
	return md.Spotify.SetArtistUrl(event)
}

func (md *syncService) SetArtistImgUrl(event *db.Event) error {
	return md.Spotify.SetArtistImgUrl(event)
}
