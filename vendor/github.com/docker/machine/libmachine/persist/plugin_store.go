package persist

import (
	"github.com/docker/machine/drivers/errdriver"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin/localbinary"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/host"
)

type PluginDriverFactory interface {
	NewPluginDriver(driverName string, rawDriver []byte) (drivers.Driver, error)
}

type RPCPluginDriverFactory struct{}

type PluginStore struct {
	*Filestore
	PluginDriverFactory
}

func (factory RPCPluginDriverFactory) NewPluginDriver(driverName string, rawDriver []byte) (drivers.Driver, error) {
	d, err := rpcdriver.NewRPCClientDriver(driverName, rawDriver)
	if err != nil {
		// Not being able to find a driver binary is a "known error"
		if _, ok := err.(localbinary.ErrPluginBinaryNotFound); ok {
			return errdriver.NewDriver(driverName), nil
		}
		return nil, err
	}

	if driverName == "virtualbox" {
		return drivers.NewSerialDriver(d), nil
	}

	return d, nil
}

func NewPluginStore(path, caCertPath, caPrivateKeyPath string) *PluginStore {
	return &PluginStore{
		Filestore:           NewFilestore(path, caCertPath, caPrivateKeyPath),
		PluginDriverFactory: RPCPluginDriverFactory{},
	}
}

func (ps PluginStore) Load(name string) (*host.Host, error) {
	h, err := ps.Filestore.Load(name)
	if err != nil {
		return nil, err
	}

	d, err := ps.NewPluginDriver(h.DriverName, h.RawDriver)
	if err != nil {
		return nil, err
	}

	h.Driver = d

	return h, nil
}
