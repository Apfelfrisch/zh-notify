package collect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractArtist(t *testing.T) {
	var tests = []struct {
		name     string
		expected string
	}{
		{"The Hirsch Effekt", "The Hirsch Effekt"},
		{"Tito & Tarantula", "Tito & Tarantula"},
		{"Bosse (Ausverkauft)", "Bosse"},
		{"Monchi (ausverkauft)", "Monchi"},
		{"Senta (Kultur für Kinder)", "Senta"},
		{"Dennis & Jesko Band (Wumms)", "Dennis & Jesko Band"},
		{"Mina Richman & Band", "Mina Richman"},
		{"Heldmaschine – Eiszeit Tour 2026", "Heldmaschine"},
		{"De Beidn – „Twee as Bonnie & Clyde“", "De Beidn"},
		{"Helene Bockhorst „Lebefrau“", "Helene Bockhorst"},
		{"Snake Eyes + schluma.", "Snake Eyes"},
		{"LACK + False Lefty", "LACK"},
		{"My´Tallica", "My´Tallica"},
		{"Live-Hörspiel: Rache zeugt die schönsten Morde", "Live-Hörspiel: Rache zeugt die schönsten Morde"},
		{"", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, extractArtist(test.name))
		})
	}
}

func TestCategoryFromClasses(t *testing.T) {
	var tests = []struct {
		classes  string
		expected string
	}{
		{"elementor e-loop-item post-7954 event hentry event_type-konzerte", "concert"},
		{"hentry event_type-comedy event_type-konzerte", "concert"},
		{"hentry event_type-konzerte event_type-party", "concert"},
		{"hentry event_type-comedy event_type-theater", "comedy"},
		{"hentry event_type-comedy event_type-sonstige", "comedy"},
		{"hentry event_type-kids event_type-theater", "theatre"},
		{"hentry event_type-party", "party"},
		{"hentry event_type-featured event_type-sonstige", "unknown"},
		{"", "unknown"},
	}

	for _, test := range tests {
		t.Run(test.classes, func(t *testing.T) {
			assert.Equal(t, test.expected, categoryFromClasses(test.classes))
		})
	}
}

func TestCleanPlace(t *testing.T) {
	var tests = []struct {
		place    string
		expected string
	}{
		{"Café →", "Café"},
		{"Theater an der Blinke →", "Theater an der Blinke"},
		{"Großer Saal\u00a0→", "Großer Saal"},
		{"Theater", "Theater"},
		{"", ""},
	}

	for _, test := range tests {
		t.Run(test.place, func(t *testing.T) {
			assert.Equal(t, test.expected, cleanPlace(test.place))
		})
	}
}
