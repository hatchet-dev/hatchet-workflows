package temporalclient

import "github.com/spf13/viper"

type TemporalClientConfigFile struct {
	// Temporal config options
	TemporalHostPort  string `mapstructure:"hostPort" json:"hostPort,omitempty" default:"127.0.0.1:7233"`
	TemporalNamespace string `mapstructure:"namespace" json:"namespace,omitempty" default:"default"`

	// TLS options
	TemporalClientTLSRootCA     string `mapstructure:"tlsRootCA" json:"tlsRootCA,omitempty"`
	TemporalClientTLSRootCAFile string `mapstructure:"tlsRootCAFile" json:"tlsRootCAFile,omitempty"`
	TemporalClientTLSCert       string `mapstructure:"tlsCert" json:"tlsCert,omitempty"`
	TemporalClientTLSCertFile   string `mapstructure:"tlsCertFile" json:"tlsCertFile,omitempty"`
	TemporalClientTLSKey        string `mapstructure:"tlsKey" json:"tlsKey,omitempty"`
	TemporalClientTLSKeyFile    string `mapstructure:"tlsKeyFile" json:"tlsKeyFile,omitempty"`
	TemporalTLSServerName       string `mapstructure:"tlsServerName" json:"tlsServerName,omitempty"`
}

func BindAllEnv(v *viper.Viper) {
	v.BindEnv("hostPort", "TEMPORAL_CLIENT_HOST_PORT")
	v.BindEnv("namespace", "TEMPORAL_CLIENT_NAMESPACE")

	v.BindEnv("tlsRootCA", "TEMPORAL_CLIENT_TLS_ROOT_CA")
	v.BindEnv("tlsRootCAFile", "TEMPORAL_CLIENT_TLS_ROOT_CA_FILE")
	v.BindEnv("tlsCert", "TEMPORAL_CLIENT_TLS_CERT")
	v.BindEnv("tlsCertFile", "TEMPORAL_CLIENT_TLS_CERT_FILE")
	v.BindEnv("tlsKey", "TEMPORAL_CLIENT_TLS_KEY")
	v.BindEnv("tlsKeyFile", "TEMPORAL_CLIENT_TLS_KEY_FILE")
	v.BindEnv("tlsServerName", "TEMPORAL_CLIENT_TLS_SERVER_NAME")
}
