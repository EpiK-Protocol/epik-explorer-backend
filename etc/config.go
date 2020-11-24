package etc

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

//config ...
type config struct {
	Server   Server   `yaml:"server"`
	BadgerDB BadgerDB `yaml:"badgerdb"`
	EPIK     EPIK     `yaml:"epik"`
}

//Server ...
type Server struct {
	Mode     string `yaml:"mode"`
	Name     string `yaml:"name"`
	HTTPPort int64  `yaml:"http_port"`
	LogDir   string `yaml:"log_dir"`
}

//EPIK ...
type EPIK struct {
	MainPrivateKey string `yaml:"main_privatekey"`
	RPCHost        string `yaml:"rpc_host"`
	RPCToken       string `yaml:"rpc_token"`
}

//BadgerDB ...
type BadgerDB struct {
	TipSet  string `yaml:"tipset"`
	Power   string `yaml:"power"`
	Message string `yaml:"message"`
}

//Config ...
var Config config

//Load ...
func Load(file string) (err error) {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bs, &Config)
}
