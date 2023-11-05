// Adapted from: https://github.com/hatchet-dev/hatchet/blob/3c2c13168afa1af68d4baaf5ed02c9d49c5f0323/internal/temporal/server/ui.go#L1

package server

import (
	"fmt"

	hatchetconfig "github.com/hatchet-dev/hatchet-workflows/internal/temporal/server/config"
	uiserver "github.com/temporalio/ui-server/v2/server"
	uiconfig "github.com/temporalio/ui-server/v2/server/config"
	uiserveroptions "github.com/temporalio/ui-server/v2/server/server_options"
)

func NewUIServer(configfile *hatchetconfig.TemporalConfigFile) (*uiserver.Server, error) {
	cfg := &uiconfig.Config{
		Host:                configfile.UI.TemporalUIAddress,
		Port:                int(configfile.UI.TemporalUIPort),
		TemporalGRPCAddress: fmt.Sprintf("%s:%d", configfile.TemporalBroadcastAddress, configfile.Frontend.TemporalFrontendPort),
		EnableUI:            configfile.UI.TemporalUIEnabled,
		TLS: uiconfig.TLS{
			CaFile:     configfile.UI.TemporalUITLSRootCAFile,
			CertFile:   configfile.UI.TemporalUITLSCertFile,
			KeyFile:    configfile.UI.TemporalUITLSKeyFile,
			ServerName: configfile.UI.TemporalUITLSServerName,
		},
	}

	return uiserver.NewServer(uiserveroptions.WithConfigProvider(cfg)), nil
}
