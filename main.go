package main

import (
	"flag"
	"os"
	"time"

	"github.com/stnokott/r6api/api/types/stats"

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

	stats := new(stats.MapStats)
	if err = a.GetStats(profile, "Y7S3", stats); err != nil {
		logger.Fatal().Err(err).Msgf("error getting summarized stats for <%s>", profile.Name)
	}
}
