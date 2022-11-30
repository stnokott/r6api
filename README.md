# r6api
## Disclaimer
- I don't have the time to maintain this repo properly.
- This project is just for fun.
- I am a Golang beginner, be kind
- Feel free to fork and do whatever you like with it.
- **I didn't have time to write tests, so use with caution!**

## Background information
- reverse-engineered official Ubisoft API at https://www.ubisoft.com/de-de/game/rainbow-six/siege/stats
- in parts inspired by [danielwerg/r6api.js](https://github.com/danielwerg/r6api.js)

## Example usage

```go
import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/stnokott/r6api"
)

func main() {
	// setup
	writer := zerolog.ConsoleWriter{
		Out:           os.Stdout,
		TimeFormat:    time.RFC3339,
		PartsOrder:    []string{"time", "level", "name", "message"},
		FieldsExclude: []string{"name"},
	}
	logger := zerolog.New(writer).Level(zerolog.DebugLevel).With().Timestamp().Str("name", "R6API -").Logger()

	email := "<myubiemail>"
	password := "<myubipassword>"

	// create API instance
	a := r6api.NewR6API(email, password, logger)

	// resolve username to profile
	profile, err := a.ResolveUser("<myubiusername>")
	if err != nil {
		logger.Fatal().Err(err).Msg("error resolving user")
	}

	// get ranked history for profile with history depth of 2
	ranked, err := a.GetRankedHistory(profile, 2)
	if err != nil {
		logger.Fatal().Err(err).Msg("error getting ranked history")
	}
	// get most-recent ranked season
	r := ranked[1]

	metadata, err := a.GetMetadata()
	if err != nil {
		logger.Fatal().Err(err).Msg("error getting metadata")
	}

	// retrieve season slug for last ranked season
	seasonSlug := metadata.SlugFromID(r.SeasonID)
	if seasonSlug == "" {
		seasonSlug = "n/a"
	}
	// print info
	logger.Info().Str("season", seasonSlug).Int("kills", r.Kills).Int("deaths", r.Deaths).Send()
}
```
