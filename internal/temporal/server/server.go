// Adapted from: https://github.com/hatchet-dev/hatchet/blob/3c2c13168afa1af68d4baaf5ed02c9d49c5f0323/internal/temporal/server/server.go#L1

package server

import (
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/temporal"

	"github.com/hatchet-dev/hatchet-workflows/internal/temporal/server/authorizers"
	hatchetconfig "github.com/hatchet-dev/hatchet-workflows/internal/temporal/server/config"
)

func NewTemporalServer(tconfig *hatchetconfig.Config, interruptCh <-chan interface{}) (temporal.Server, error) {
	configfile := tconfig.ConfigFile

	logger := log.NewZapLogger(log.BuildZapLogger(log.Config{
		Stdout:     true,
		Level:      configfile.TemporalLogLevel,
		OutputFile: "",
	}))

	cfg, err := GetTemporalServerConfig(tconfig)
	if err != nil {
		return nil, err
	}

	authorizerAndClaimMapper := authorizers.NewCertificateAuthorizer(tconfig, &cfg.Global.Authorization, logger)

	return temporal.NewServer(
		temporal.ForServices(temporal.DefaultServices),
		temporal.WithConfig(cfg),
		temporal.WithLogger(logger),
		temporal.InterruptOn(interruptCh),
		temporal.WithAuthorizer(authorizerAndClaimMapper),
		temporal.WithClaimMapper(func(cfg *config.Config) authorization.ClaimMapper {
			return authorizerAndClaimMapper
		}),
	)
}
