package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"go.temporal.io/sdk/client"
)

const HatchetDefaultQueueName = "default"

type ClientOptions struct {
	DefaultQueueName string

	HostPort  string
	Namespace string

	ClientKey     []byte
	ClientKeyFile string

	ClientCert     []byte
	ClientCertFile string

	RootCA     []byte
	RootCAFile string

	TLSServerName string
}

func defaultOpts() *ClientOptions {
	return &ClientOptions{
		DefaultQueueName: HatchetDefaultQueueName,
	}
}

func WithDefaultQueueName(queueName string) ClientOpt {
	return func(workerOptions *ClientOptions) {
		workerOptions.DefaultQueueName = queueName
	}
}

func WithCerts(clientCert, clientKey []byte) ClientOpt {
	return func(workerOptions *ClientOptions) {
		workerOptions.ClientCert = clientCert
		workerOptions.ClientKey = clientKey
	}
}

func WithCertFiles(clientCertFile, clientKeyFile string) ClientOpt {
	return func(workerOptions *ClientOptions) {
		workerOptions.ClientCertFile = clientCertFile
		workerOptions.ClientKeyFile = clientKeyFile
	}
}

func WithRootCA(rootCA []byte) ClientOpt {
	return func(workerOptions *ClientOptions) {
		workerOptions.RootCA = rootCA
	}
}

func WithRootCAFile(rootCAFile string) ClientOpt {
	return func(workerOptions *ClientOptions) {
		workerOptions.RootCAFile = rootCAFile
	}
}

func WithTLSServerName(tlsServerName string) ClientOpt {
	return func(workerOptions *ClientOptions) {
		workerOptions.TLSServerName = tlsServerName
	}
}

func WithHostPort(hostPort string) ClientOpt {
	return func(workerOptions *ClientOptions) {
		workerOptions.HostPort = hostPort
	}
}

func WithNamespace(namespace string) ClientOpt {
	return func(workerOptions *ClientOptions) {
		workerOptions.Namespace = namespace
	}
}

type ClientOpt func(*ClientOptions)

type Client struct {
	opts *ClientOptions

	clients sync.Map
}

func NewTemporalClient(opts ...ClientOpt) (*Client, error) {
	options := defaultOpts()

	for _, opt := range opts {
		opt(options)
	}

	res := &Client{options, sync.Map{}}

	err := res.eventualClientFromOpts(options, options.DefaultQueueName, 5)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) eventualClientFromOpts(opts *ClientOptions, taskQueueName string, maxRetries uint) error {
	tOpts := client.Options{
		HostPort:  opts.HostPort,
		Namespace: opts.Namespace,
		Identity:  fmt.Sprintf("%d@%s@%s", os.Getpid(), getHostName(), taskQueueName),
	}

	if (opts.ClientCert != nil || opts.ClientCertFile != "") && (opts.ClientKey != nil || opts.ClientKeyFile != "") {
		tlsConfig := &tls.Config{
			ServerName: opts.TLSServerName,
			MinVersion: tls.VersionTLS12,
		}

		var cert tls.Certificate
		var err error

		if opts.ClientCert != nil && opts.ClientKey != nil {
			cert, err = tls.X509KeyPair(opts.ClientCert, opts.ClientKey)

			if err != nil {
				return fmt.Errorf("unable to load client cert and key from env: %v", err)
			}
		} else if opts.ClientCertFile != "" && opts.ClientKeyFile != "" {
			cert, err = tls.LoadX509KeyPair(
				opts.ClientCertFile,
				opts.ClientKeyFile,
			)

			if err != nil {
				return fmt.Errorf("unable to load client cert and key from files: %v", err)
			}
		} else {
			return errors.New("specify both client cert and key as files or both as environment variables")
		}

		caPool := x509.NewCertPool()
		var caBytes []byte

		caBytes, err = os.ReadFile(
			opts.RootCAFile,
		)

		if err != nil {
			return fmt.Errorf("unable to load CA cert from file: %v", err)
		}

		if !caPool.AppendCertsFromPEM(caBytes) {
			return errors.New("unknown failure constructing cert pool for ca")
		}

		tlsConfig.RootCAs = caPool

		tlsConfig.Certificates = []tls.Certificate{cert}

		tOpts.ConnectionOptions.TLS = tlsConfig
	}

	getter := func() {
		var err error
		var tClient client.Client

		for i := 0; i < int(maxRetries); i++ {
			tClient, err = client.Dial(tOpts)

			if err == nil {
				c.clients.Store(taskQueueName, tClient)
				break
			} else {
				fmt.Fprintf(os.Stderr, fmt.Sprintf("could not create temporal client for queue %s: %s. Retrying (attempt %d of %d)...\n", taskQueueName, err.Error(), i+1, maxRetries))
				time.Sleep(5 * time.Second)
			}
		}

		if err != nil {
			// TODO: use shared logger here
			fmt.Fprintf(os.Stderr, fmt.Sprintf("Fatal: could not create temporal client for queue %s: %s\n", taskQueueName, err.Error()))
		}
	}

	if maxRetries == 1 {
		getter()
	} else {
		go getter()
	}

	return nil
}

func (c *Client) GetClient(queueName string) (client.Client, error) {
	if queueName == "" {
		if res, ok := c.clients.Load(c.opts.DefaultQueueName); ok {
			return res.(client.Client), nil
		}
	}

	tc, exists := c.clients.Load(queueName)

	if !exists {
		return c.newQueueClient(queueName)
	}

	return tc.(client.Client), nil
}

func (c *Client) GetDefaultQueueName() string {
	return c.opts.DefaultQueueName
}

func (c *Client) newQueueClient(taskQueueName string) (client.Client, error) {
	err := c.eventualClientFromOpts(c.opts, taskQueueName, 1)
	if err != nil {
		return nil, err
	}

	res, ok := c.clients.Load(taskQueueName)

	if !ok {
		return nil, fmt.Errorf("could not load temporal client for queue %s", taskQueueName)
	}

	return res.(client.Client), nil
}

func (c *Client) Close() {
	c.Close()
}

func getHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "Unknown"
	}
	return hostName
}
