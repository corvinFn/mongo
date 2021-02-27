package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/pkg/errors"
)

var (
	Gdc *Client // golden-cloud
)

func Init(env string) error {
	var conf map[string][]string

	switch env {
	case "local", "dev":
		conf = DevCnf
	case "test":
		conf = TestCnf
	case "prod":
		conf = ProdCnf
	default:
		return errors.New("illegal env")
	}

	return InitClients(conf)
}

// InitClients init clients with config
func InitClients(conf map[string][]string) error {
	for name, adds := range conf {
		client := &Client{
			addresses: adds,
			mode:      mgo.Eventual,
		}

		switch name {
		case "gdc":
			Gdc = client
		default:
			return errors.Errorf("Unknown db shortcut: '%s'", name)
		}
	}

	return nil
}
