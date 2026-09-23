package collect

import (
	"regexp"
	"slices"
	"strings"
)

var categoryByEventType = []struct {
	eventType string
	category  string
}{
	{"event_type-konzerte", "concert"},
	{"event_type-comedy", "comedy"},
	{"event_type-theater", "theatre"},
	{"event_type-party", "party"},
}

func categoryFromClasses(classes string) string {
	fields := strings.Fields(classes)

	for _, m := range categoryByEventType {
		if slices.Contains(fields, m.eventType) {
			return m.category
		}
	}

	return "unknown"
}

var (
	parenthesesRe = regexp.MustCompile(`\([^)]*\)`)
	quotedRe      = regexp.MustCompile(`[„“"][^“”"]*[“”"]`)
	subtitleRe    = regexp.MustCompile(`\s+[–—+-]\s+.*$`)
	bandSuffixRe  = regexp.MustCompile(`(?i)\s*&\s*band$`)
	whitespaceRe  = regexp.MustCompile(`\s+`)
)

func extractArtist(name string) string {
	artist := parenthesesRe.ReplaceAllString(name, "")
	artist = quotedRe.ReplaceAllString(artist, "")
	artist = subtitleRe.ReplaceAllString(artist, "")
	artist = whitespaceRe.ReplaceAllString(artist, " ")
	artist = bandSuffixRe.ReplaceAllString(strings.TrimSpace(artist), "")

	return strings.TrimSpace(artist)
}

func cleanPlace(place string) string {
	return strings.TrimSpace(strings.TrimRight(place, "→ \u00a0"))
}
