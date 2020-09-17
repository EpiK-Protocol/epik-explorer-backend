package etc

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

//config ...
type config struct {
	Server    Server       `yaml:"server"`
	Postgres  Postgres     `yaml:"postgres"`
	BadgerDB  BadgerDB     `yaml:"badgerdb"`
	EPIK      EPIK         `yaml:"epik"`
	EPIKERC20 EPIKERC20    `yaml:"epik_erc20"`
	Wallet    WalletConfig `yaml:"wallet"`
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

//EPIKERC20 ...
type EPIKERC20 struct {
	MainPrivateKey string  `yaml:"main_privatekey"`
	RPCHost        string  `yaml:"rpc_host"`
	DailyBonus     float64 `yaml:"daily_bonus"`
}

//Postgres ...
type Postgres struct {
	Host     string `yaml:"host"`
	Port     int64  `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

//BadgerDB ...
type BadgerDB struct {
	TipSet  string `yaml:"tipset"`
	Power   string `yaml:"power"`
	Message string `yaml:"message"`
	TestNet string `yaml:"testnet"`
	Wallet  string `yaml:"wallet"`
	Token   string `yaml:"token"`
}

//WalletConfig ...
type WalletConfig struct {
	Android   ClientVersion `yaml:"android"`
	IOS       ClientVersion `yaml:"ios"`
	WalletAPI string        `yaml:"wallet_api"`
	ETHAPI    string        `yaml:"eth_api"`
	EPKAPI    string        `yaml:"epk_api"`
	EPKToken  string        `yaml:"epk_token"`
}

//ClientVersion ...
type ClientVersion struct {
	LaestVersion    string `yaml:"laest_version"`
	RequiredVersion string `yaml:"required_version"`
	UpdateURL       string `yaml:"update_url"`
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
