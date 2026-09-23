package internal

import (
	"github.com/apfelfrisch/zh-notify/internal/collect"
	"github.com/apfelfrisch/zh-notify/internal/collect/spotify"
	"github.com/apfelfrisch/zh-notify/internal/db"
)

const URL = "https://www.zollhaus-leer.com/veranstaltungen/"

func CollectNewEvents() ([]collect.Event, error) {
	return collect.CrawlEvents(URL)
}

func NewSyncEventCollector(spotifyId, sporitySecret string) *SyncCollector {
	return &SyncCollector{
		service: &syncService{
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
	Spotify *spotify.Service
}

func (md *syncService) Init() error {
	return md.Spotify.Init()
}

func (md *syncService) SetArtistUrl(event *db.Event) error {
	return md.Spotify.SetArtistUrl(event)
}

func (md *syncService) SetArtistImgUrl(event *db.Event) error {
	return md.Spotify.SetArtistImgUrl(event)
}
