package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	// General initialization.
	rand.Seed(time.Now().UnixNano())
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("WITH_DEBUG") == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	lg := zerolog.New(os.Stdout).With().Timestamp().Logger()

	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		lg.Fatal().Msg("provide a command, either 'server' or 'e2e'")
	}

	switch args[0] {
	case "server":
		runServerCommand(lg)
	case "e2e":
		runE2eCommand(lg)
	default:
		lg.Fatal().Msg(fmt.Sprintf("invalid command '%s' provided", args[0]))
	}
}
