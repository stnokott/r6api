package main

import (
	"flag"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/stnokott/r6api/api"
)

func main() {
	writer := zerolog.ConsoleWriter{
		Out:           os.Stdout,
		TimeFormat:    time.RFC3339,
		PartsOrder:    []string{"time", "level", "name", "message"},
		FieldsExclude: []string{"name"},
	}
	logger := zerolog.New(writer).Level(zerolog.DebugLevel).With().Timestamp().Str("name", "UbiAPI -").Logger()

	email := flag.String("email", "", "Ubisoft email")
	password := flag.String("pass", "", "Ubisoft password")
	flag.Parse()
	if *email == "" || *password == "" {
		flag.Usage()
		logger.Fatal().Msg("missing flags")
	}

	a := api.NewUbiAPI(*email, *password, logger)
	profile, err := a.ResolveUser("Knoblauch.SOOS")
	if err != nil {
		logger.Fatal().Err(err).Msg("error resolving user")
	}

	metadata, err := a.GetMetadata()
	if err != nil {
		logger.Fatal().Err(err).Msg("error getting metadata")
	}

	ranked, err := a.GetRankedHistory(profile, 1)
	if err != nil {
		logger.Fatal().Err(err).Msg("error getting ranked history")
	}
	r := ranked[0]
	seasonSlug := metadata.SlugFromID(r.SeasonID)
	if seasonSlug == "" {
		seasonSlug = "n/a"
	}
	logger.Info().Str("season", seasonSlug).Int("kills", r.Kills).Int("deaths", r.Deaths).Send()
}
