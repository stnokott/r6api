package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/stnokott/r6api"
	"github.com/stnokott/r6api/types/stats"
)

func main() {
	writer := zerolog.ConsoleWriter{
		Out:           os.Stdout,
		TimeFormat:    time.RFC3339,
		PartsOrder:    []string{"time", "level", "name", "message"},
		FieldsExclude: []string{"name"},
	}
	logger := zerolog.New(writer).Level(zerolog.DebugLevel).With().Timestamp().Str("name", "R6API").Logger()

	a := r6api.NewR6API("checker13579@gmail.com", "wj5fvonfZX", logger)
	profile, _ := a.ResolveUser("Knoblauch.SOOS")

	m, err := a.GetMetadata()
	if err != nil {
		panic(err)
	}
	logger.Info().Str("seasonSlug", m.Seasons[len(m.Seasons)-1].Slug).Str("seasonName", m.Seasons[len(m.Seasons)-1].Name).Msg("current season")

	for i := 0; i < 2; i++ {
		seasonSlug := m.Seasons[len(m.Seasons)-1-i].Slug
		stats := new(stats.MapStats)
		if err := a.GetStats(profile, seasonSlug, stats); err != nil {
			panic(err)
		}

		// Summarized
		//logger.Info().Str("season", stats.SeasonSlug).Int("kills", stats.All.All.Kills).Send()
		// Maps
		logger.Info().Str("season", stats.SeasonSlug).Int("matches_played", (*stats.All)["BANK V2"].MatchesPlayed).Send()
		// Operator
		//logger.Info().Str("season", stats.SeasonSlug).Str("operator_name", "all").Int("kills", stats.All.All["all"].Kills).Send()
		// Weapons
		//logger.Info().Str("season", seasonSlug).Send()
	}

	skillHistory, err := a.GetRankedHistory(profile, 1)
	if err != nil {
		panic(err)
	}
	rankedStats := skillHistory[0]
	logger.Info().Str("season", m.SeasonNameFromID(rankedStats.SeasonID)).Msg("got ranked stats")
}
