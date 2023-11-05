// Adapted from: https://github.com/hatchet-dev/hatchet/blob/3c2c13168afa1af68d4baaf5ed02c9d49c5f0323/internal/config/loader/loader.go

package loader

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/creasty/defaults"
	"github.com/spf13/viper"

	temporalconfig "github.com/hatchet-dev/hatchet-workflows/internal/temporal/server/config"
	"github.com/hatchet-dev/hatchet-workflows/pkg/client"
	clientconfig "github.com/hatchet-dev/hatchet-workflows/pkg/client/config"
)

// LoadTemporalClient loads the temporal client via viper
func LoadTemporalClientConfigFile(files ...[]byte) (*clientconfig.TemporalClientConfigFile, error) {
	configFile := &clientconfig.TemporalClientConfigFile{}
	f := clientconfig.BindAllEnv

	_, err := loadConfigFromViper(f, configFile, files...)

	return configFile, err
}

// LoadTemporalConfigFile loads the temporal config file via viper
func LoadTemporalConfigFile(files ...[]byte) (*temporalconfig.TemporalConfigFile, error) {
	configFile := &temporalconfig.TemporalConfigFile{}
	f := temporalconfig.BindAllEnv

	_, err := loadConfigFromViper(f, configFile, files...)

	return configFile, err
}

func loadConfigFromViper(bindFunc func(v *viper.Viper), configFile interface{}, files ...[]byte) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	bindFunc(v)

	for _, f := range files {
		err := v.MergeConfig(bytes.NewBuffer(f))

		if err != nil {
			return nil, fmt.Errorf("could not load viper config: %w", err)
		}
	}

	defaults.Set(configFile)

	err := v.Unmarshal(configFile)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal viper config: %w", err)
	}

	return v, nil
}

type ConfigLoader struct {
	version, directory string
}

// LoadTemporalClientConfig loads the temporal client configuration
func (c *ConfigLoader) LoadTemporalClient() (res *client.Client, err error) {
	sharedFilePath := filepath.Join(c.directory, "temporal-client.yaml")

	configFileBytes, err := getConfigBytes(sharedFilePath)

	if err != nil {
		return nil, err
	}

	cf, err := LoadTemporalClientConfigFile(configFileBytes...)

	if err != nil {
		return nil, err
	}

	return GetTemporalClientFromConfigFile(cf)
}

// LoadTemporalConfig loads the temporal server configuration
func (c *ConfigLoader) LoadTemporalConfig() (res *temporalconfig.Config, err error) {
	sharedFilePath := filepath.Join(c.directory, "temporal.yaml")
	configFileBytes, err := getConfigBytes(sharedFilePath)

	if err != nil {
		return nil, err
	}

	cf, err := LoadTemporalConfigFile(configFileBytes...)

	if err != nil {
		return nil, err
	}

	return GetTemporalConfigFromConfigFile(cf)
}

func getConfigBytes(configFilePath string) ([][]byte, error) {
	configFileBytes := make([][]byte, 0)

	if fileExists(configFilePath) {
		fileBytes, err := ioutil.ReadFile(configFilePath) // #nosec G304 -- config files are meant to be read from user-supplied directory

		if err != nil {
			return nil, fmt.Errorf("could not read config file at path %s: %w", configFilePath, err)
		}

		configFileBytes = append(configFileBytes, fileBytes)
	}

	return configFileBytes, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func GetTemporalConfigFromConfigFile(
	tc *temporalconfig.TemporalConfigFile,
) (res *temporalconfig.Config, err error) {
	return &temporalconfig.Config{
		ConfigFile: tc,
	}, nil
}

func GetTemporalClientFromConfigFile(
	tc *clientconfig.TemporalClientConfigFile,
) (res *client.Client, err error) {
	opts := []client.ClientOpt{}

	if tc.TemporalClientTLSCertFile != "" && tc.TemporalClientTLSKeyFile != "" {
		opts = append(opts, client.WithCertFiles(tc.TemporalClientTLSCertFile, tc.TemporalClientTLSKeyFile))
	}

	if tc.TemporalClientTLSCert != "" && tc.TemporalClientTLSKey != "" {
		opts = append(opts, client.WithCerts([]byte(tc.TemporalClientTLSCert), []byte(tc.TemporalClientTLSKey)))
	}

	if tc.TemporalClientTLSRootCAFile != "" {
		opts = append(opts, client.WithRootCAFile(tc.TemporalClientTLSRootCAFile))
	}

	if tc.TemporalClientTLSRootCA != "" {
		opts = append(opts, client.WithRootCA([]byte(tc.TemporalClientTLSRootCA)))
	}

	if tc.TemporalTLSServerName != "" {
		opts = append(opts, client.WithTLSServerName(tc.TemporalTLSServerName))
	}

	if tc.TemporalHostPort != "" {
		opts = append(opts, client.WithHostPort(tc.TemporalHostPort))
	}

	if tc.TemporalNamespace != "" {
		opts = append(opts, client.WithNamespace(tc.TemporalNamespace))
	}

	c, err := client.NewTemporalClient(
		opts...,
	)

	if err != nil {

		return nil, fmt.Errorf("could not create temporal client: %w", err)
	}

	return c, nil
}
