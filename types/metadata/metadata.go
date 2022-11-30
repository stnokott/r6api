package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
)

const URL string = "https://www.ubisoft.com/de-de/game/rainbow-six/siege/stats/glossary/fa94e165-6328-4a9b-8581-81735ffaba27"

// New creates a new instance, parsing scriptJS.
// The parameter should be a Javascript string starting with "window.__PRELOADED_STATE__ = <JS object>".
// This method should only be called internally.
func New(scriptJS string) (*Metadata, error) {
	vm := otto.New()
	if _, err := vm.Run(strings.Replace(scriptJS, "window.__PRELOADED_STATE__", "state", 1)); err != nil {
		return nil, fmt.Errorf("could not run JS in VM: %w", err)
	}
	state, err := vm.Get("state")
	if err != nil {
		return nil, fmt.Errorf("could not get VM variable value: %w", err)
	}
	stateJSON, err := state.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("could not marshal JS data to JSON: %w", err)
	}

	m := new(Metadata)
	if err := json.Unmarshal(stateJSON, m); err != nil {
		return nil, err
	}

	return m, nil
}

type Metadata struct {
	Seasons []season
}

type season struct {
	Name      string
	Slug      string
	StartDate time.Time
	Ranks     []rank
}

type rank struct {
	MinMMR int    `json:"min"`
	MaxMMR int    `json:"max"`
	Slug   string `json:"slug"`
}

type rootJSON struct {
	ContentfulGraphQl map[string]json.RawMessage `json:"ContentfulGraphQl"`
}

type seasonsJSON struct {
	Content struct {
		Seasons []struct {
			Slug           string `json:"slug"`
			LocalizedItems struct {
				Title string `json:"title"`
			} `json:"localizedItems"`
			StartDate time.Time `json:"startDate"`
			RankList  struct {
				Data struct {
					Ranks []rank `json:"ranks"`
				} `json:"data"`
			} `json:"rankList"`
		} `json:"seasons"`
	} `json:"content"`
}

func (m *Metadata) UnmarshalJSON(data []byte) error {
	var root rootJSON

	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}

	for k, v := range root.ContentfulGraphQl {
		if strings.HasPrefix(k, "G2W Card-") {
			var seasonsJSON seasonsJSON
			if err := json.Unmarshal(v, &seasonsJSON); err != nil {
				return err
			}
			seasons := make([]season, len(seasonsJSON.Content.Seasons))
			for i, seasonJSON := range seasonsJSON.Content.Seasons {
				seasons[i] = season{
					Name:      seasonJSON.LocalizedItems.Title,
					Slug:      seasonJSON.Slug,
					StartDate: seasonJSON.StartDate,
					Ranks:     seasonJSON.RankList.Data.Ranks,
				}
			}
			m.Seasons = seasons
			return nil
		}
	}
	return errors.New("could not find required key in data")
}

// SeasonSlugFromID will return the slug of the season (e.g. "Y7S3") with the provided ID or "" if unknown.
func (m *Metadata) SeasonSlugFromID(seasonID int) string {
	if len(m.Seasons) < seasonID-1 {
		return ""
	} else {
		return m.Seasons[seasonID-1].Slug
	}
}

// SeasonNameFromID will return the name of the season (e.g. "Brutal Swarm") with the provided ID or "" if unknown.
func (m *Metadata) SeasonNameFromID(seasonID int) string {
	if len(m.Seasons) < seasonID-1 {
		return ""
	} else {
		return m.Seasons[seasonID-1].Name
	}
}
