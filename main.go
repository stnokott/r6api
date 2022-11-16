package main

import (
	"flag"
	"github.com/stnokott/r6api/api/types/stats"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/stnokott/r6api/api"
)

func main() {
	email := flag.String("email", "", "Ubisoft email")
	password := flag.String("pass", "", "Ubisoft password")
	flag.Parse()

	writer := zerolog.ConsoleWriter{
		Out:           os.Stdout,
		TimeFormat:    time.RFC3339,
		PartsOrder:    []string{"time", "level", "name", "message"},
		FieldsExclude: []string{"name"},
	}
	logger := zerolog.New(writer).Level(zerolog.DebugLevel).With().Timestamp().Str("name", "UbiAPI -").Logger()

	a := api.NewUbiAPI(*email, *password, logger)
	profile, err := a.ResolveUser("Knoblauch.SOOS")
	if err != nil {
		logger.Fatal().Err(err).Msg("error resolving user")
	}

	summarizedStats := new(stats.SummarizedStats)
	if err = a.GetStats(profile, "Y7S3", summarizedStats); err != nil {
		logger.Fatal().Err(err).Msgf("error getting summarized stats for <%s>", profile.Name)
	}

	operatorStats := new(stats.OperatorStats)
	if err = a.GetStats(profile, "Y7S3", operatorStats); err != nil {
		logger.Fatal().Err(err).Msgf("error getting operator stats for <%s>", profile.Name)
	}
	if operatorStats.Casual != nil {
		for _, operator := range operatorStats.Casual.Attack {
			logger.Info().
				Str("role", "attack").
				Int("kills", operator.Kills).
				Int("deaths", operator.Deaths).
				Msg(operator.Name)
		}
	}

	mapStats := new(stats.MapStats)
	if err = a.GetStats(profile, "Y7S3", mapStats); err != nil {
		logger.Fatal().Err(err).Msgf("error getting map stats for <%s>", profile.Name)
	}
	if mapStats.Casual != nil {
		for _, map_ := range mapStats.Casual.Attack {
			logger.Info().
				Str("role", "attack").
				Int("entrykills", map_.EntryKills).
				Int("entrydeaths", map_.EntryDeaths).
				Msg(map_.Name)
		}
	}

	weaponStats := new(stats.WeaponStats)
	if err = a.GetStats(profile, "Y7S3", weaponStats); err != nil {
		logger.Fatal().Err(err).Msgf("error getting weapon stats for <%s>", profile.Name)
	}
	if weaponStats.Casual != nil {
		primaryStats := weaponStats.Casual.Attack.Primary
		for weaponType, pStats := range primaryStats {
			for _, stat := range pStats {
				logger.Info().
					Str("role", "attack").
					Str("type", weaponType).
					Int("kills", stat.Kills).
					Float64("headshotperc", stat.HeadshotPercentage).
					Msg(stat.WeaponName)
			}
		}
	}

	rankedHistory, err := a.GetRankedHistory(profile, 8)
	if err != nil {
		logger.Fatal().Err(err).Msgf("error getting weapon stats for <%s>", profile.Name)
	}
	season := rankedHistory[6]
	logger.Info().
		Int("seasonID", season.SeasonID).
		Int("wins", season.Wins).
		Int("losses", season.Losses).
		Int("mmr", season.MMR).
		Msg("")
}
