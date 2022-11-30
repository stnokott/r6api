# Disclaimer
I don't have the time to maintain this repo properly.
This is just for fun.
Feel free to fork and do whatever you like with it.

## Example usage

```go
import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/stnokott/r6api" api
)

// setup
writer := zerolog.ConsoleWriter{
	Out:           os.Stdout,
	TimeFormat:    time.RFC3339,
	PartsOrder:    []string{"time", "level", "name", "message"},
	FieldsExclude: []string{"name"},
}
logger := zerolog.New(writer).Level(zerolog.DebugLevel).With().Timestamp().Str("name", "UbiAPI -").Logger()

email := "myubiemail"
password := "myubipassword"

// create API instance
a := api.NewUbiAPI(email, password, logger)

// resolve username to profile
profile, err := a.ResolveUser("MyR6Username")
if err != nil {
	logger.Fatal().Err(err).Msg("error resolving user")
}

// get ranked history for profile with history depth of 1
ranked, err := a.GetRankedHistory(profile, 1)
if err != nil {
	logger.Fatal().Err(err).Msg("error getting ranked history")
}

// retrieve season slug for last ranked season
r := ranked[0]
seasonSlug := metadata.SlugFromID(r.SeasonID)
if seasonSlug == "" {
	seasonSlug = "n/a"
}
// print info
logger.Info().Str("season", seasonSlug).Int("kills", r.Kills).Int("deaths", r.Deaths).Send()
```
