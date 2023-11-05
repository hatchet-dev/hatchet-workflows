// Adapted from: https://github.com/hatchet-dev/hatchet/blob/3c2c13168afa1af68d4baaf5ed02c9d49c5f0323/cmd/hatchet-temporal/main.go

package main

import (
	"fmt"
	goLog "log"
	"os"

	"github.com/hatchet-dev/hatchet-workflows/cmd/cmdutils"
	"github.com/hatchet-dev/hatchet-workflows/internal/config/loader"
	"github.com/hatchet-dev/hatchet-workflows/internal/temporal/server"

	// Load sqlite storage driver
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite"
)

type uiConfig struct {
	Host                string
	Port                int
	TemporalGRPCAddress string
	EnableUI            bool
	CodecEndpoint       string
}

func main() {
	configLoader := &loader.ConfigLoader{}
	interruptChan := cmdutils.InterruptChan()
	tc, err := configLoader.LoadTemporalConfig()

	if err != nil {
		fmt.Printf("Fatal: could not load server config: %v\n", err)
		os.Exit(1)
	}

	s, err := server.NewTemporalServer(tc, interruptChan)

	if err != nil {
		goLog.Fatal(err)
	}

	sui, err := server.NewUIServer(tc.ConfigFile)

	if err != nil {
		goLog.Fatal(fmt.Sprintf("Unable to create UI server. Error: %v\n", err))
	}

	go func() {
		if err := sui.Start(); err != nil {
			panic(err)
		}
	}()

	err = s.Start()

	if err != nil {
		goLog.Fatal(fmt.Sprintf("Unable to start server. Error: %v\n", err))
	}

	return
}
