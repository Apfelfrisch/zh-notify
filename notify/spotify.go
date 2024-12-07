package notify

import (
	"context"
	"database/sql"
	"math"
	"sort"

	"github.com/apfelfrisch/zh-notify/db"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
)

type spotifyResp struct {
	event  db.Event
	artist spotify.FullArtist
}

type spotifyService struct {
	auth     *clientcredentials.Config
	client   *spotify.Client
	response *spotifyResp
}

func (sp *spotifyService) Init() error {
	accessToken, err := sp.auth.Token(context.Background())
	if err != nil {
		return err
	}

	client := spotify.Authenticator{}.NewClient(accessToken)
	sp.client = &client

	return nil
}

func (sp *spotifyService) SetArtistUrl(event *db.Event) error {
	if !event.Artist.Valid || !(event.Category.String == "concert" || event.Category.String == "comedy") {
		return nil
	}

	artist, err := sp.requestArtist(event)

	if err != nil {
		return err
	}

	artistUrl := artist.ExternalURLs["spotify"]

	if artistUrl != "" {
		event.ArtistUrl = sql.NullString{String: artistUrl, Valid: true}
	}

	return nil
}

func (sp *spotifyService) SetArtistImgUrl(event *db.Event) error {
	if !event.Artist.Valid || !(event.Category.String == "concert" || event.Category.String == "comedy") {
		return nil
	}

	artist, err := sp.requestArtist(event)

	if err != nil {
		return err
	}

	artistImgUrl := filterImage(artist, 320)

	if artistImgUrl != "" {
		event.ArtistImgUrl = sql.NullString{String: artistImgUrl, Valid: true}
	}

	return nil
}

func (sp *spotifyService) requestArtist(event *db.Event) (spotify.FullArtist, error) {
	if sp.response != nil && sp.response.event.ID == event.ID && event.Artist.Valid {
		return sp.response.artist, nil
	}

	result, err := sp.client.Search("artist:"+event.Artist.String, spotify.SearchTypeArtist)

	if err != nil {
		return spotify.FullArtist{}, nil
	}

	if len(result.Artists.Artists) == 0 {
		return spotify.FullArtist{}, nil
	}

	sp.response = &spotifyResp{
		event:  *event,
		artist: result.Artists.Artists[0],
	}

	return result.Artists.Artists[0], nil
}

func filterImage(artist spotify.FullArtist, size float64) string {
	if len(artist.Images) == 0 {
		return ""
	}

	if len(artist.Images) > 1 {
		sort.Slice(artist.Images, func(i, j int) bool {
			a := math.Abs(float64(artist.Images[i].Width) - size)
			b := math.Abs(float64(artist.Images[j].Width) - size)

			return a < b
		})
	}

	return artist.Images[0].URL
}
