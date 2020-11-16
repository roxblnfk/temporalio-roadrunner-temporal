package temporal

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2"
	"github.com/spiral/roadrunner/v2/interfaces/log"
	"github.com/spiral/roadrunner/v2/plugins/config"
	rrt "github.com/temporalio/roadrunner-temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
)

const PluginName = "temporal"

type Config struct {
	Address    string
	Namespace  string
	Activities *roadrunner.PoolConfig
}

type Temporal interface {
	GetClient() (client.Client, error)
	GetDataConverter() converter.DataConverter
	GetConfig() Config
	CreateWorker(taskQueue string, options worker.Options) (worker.Worker, error)
}

// inherit roadrunner.rpc.Plugin interface
type Plugin struct {
	cfg    Config
	dc     converter.DataConverter
	log    log.Logger
	client client.Client
}

// logger dep also
func (srv *Plugin) Init(cfg config.Configurer, log log.Logger) error {
	srv.log = log
	srv.dc = rrt.NewDataConverter(converter.GetDefaultDataConverter())

	return cfg.UnmarshalKey(PluginName, &srv.cfg)
}

// GetConfig returns temporal configuration.
func (srv *Plugin) GetConfig() Config {
	return srv.cfg
}

// GetDataConverter returns data active data converter.
func (srv *Plugin) GetDataConverter() converter.DataConverter {
	return srv.dc
}

// Serve starts temporal srv.
func (srv *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	var err error

	srv.client, err = client.NewClient(client.Options{
		Logger:        srv.log,
		HostPort:      srv.cfg.Address,
		Namespace:     srv.cfg.Namespace,
		DataConverter: srv.dc,
	})

	srv.log.Debug("Connected to temporal server", "Plugin", srv.cfg.Address)

	if err != nil {
		errCh <- errors.E(errors.Op("srv connect"), err)
	}

	return errCh
}

// Stop stops temporal srv connection.
func (srv *Plugin) Stop() error {
	if srv.client != nil {
		srv.client.Close()
	}

	return nil
}

// GetClient returns active srv connection.
func (srv *Plugin) GetClient() (client.Client, error) {
	return srv.client, nil
}

// CreateWorker allocates new temporal worker on an active connection.
func (srv *Plugin) CreateWorker(tq string, options worker.Options) (worker.Worker, error) {
	if srv.client == nil {
		return nil, errors.E("unable to create worker, invalid temporal srv")
	}

	return worker.New(srv.client, tq, options), nil
}

// Name of the service.
func (srv *Plugin) Name() string {
	return PluginName
}

// RPCService returns associated rpc service.
func (srv *Plugin) RPC() interface{} {
	return &rpc{srv: srv}
}
