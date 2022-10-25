package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Server          string        `env:"ADDRESS"`
	StoreFile       string        `env:"STORE_FILE"`
	KeyFileFlag     string        `env:"KEY"`
	DatabaseEnv     string        `env:"DATABASE_DSN"`
	StoreInterval   time.Duration `env:"STORE_INTERVAL"`
	RestoreFlag     bool          `env:"RESTORE"`
	PrivateKeyFile  string        `env:"CRYPTO_KEY"`
	ConfigFile      string        `env:"CONFIG"`
	TrustedSubnet   string        `env:"TRUSTED_SUBNET"`
	CertGRPCFile    string        `env:"CERT_FILE"`
	CertKeyGRPCFile string        `env:"CERT_KEY_FILE"`
}

var (
	srvAdressArg      *string
	storeIntervalArg  *time.Duration
	storeFileArg      *string
	restoreFlagArg    *bool
	keyFileArg        *string
	databaseArg       *string
	PrivateKeyFileArg *string
	configFile        *string
	trustedSubnet     *string
	certFile          *string
	keyCertFile       *string
)

func init() {
	srvAdressArg = flag.String("a", "localhost:8080", "host:port (default localhost:8080)")
	storeIntervalArg = flag.Duration("i", time.Duration(300*time.Second), "store interval in seconds (default 300s)")
	storeFileArg = flag.String("f", "/tmp/devops-metrics-db.json", "path to file for store (default '/tmp/devops-metrics-db.json')")
	restoreFlagArg = flag.Bool("r", true, "if is true restore data from env:RESTORE (default true)")
	keyFileArg = flag.String("k", "", "hmac key")
	databaseArg = flag.String("d", "", "string database config")
	PrivateKeyFileArg = flag.String("crypto-key", "", "private key")
	configFile = flag.String("c", "", "json config file path")
	configFile = flag.String("config", "", "json config file path")
	trustedSubnet = flag.String("t", "", "trusted subnet")
	certFile = flag.String("cert", "", "tls cert file path for gRPC")
	keyCertFile = flag.String("cert-key", "", "tls key for cert file path for gRPC")
}
func parseFlag() Config {
	flag.Parse()
	return Config{
		Server:          *srvAdressArg,
		StoreFile:       *storeFileArg,
		KeyFileFlag:     *keyFileArg,
		DatabaseEnv:     *databaseArg,
		StoreInterval:   *storeIntervalArg,
		RestoreFlag:     *restoreFlagArg,
		PrivateKeyFile:  *PrivateKeyFileArg,
		ConfigFile:      *configFile,
		TrustedSubnet:   *trustedSubnet,
		CertGRPCFile:    *certFile,
		CertKeyGRPCFile: *keyCertFile,
	}
}
func (c *Config) isDefault(flagName string, envName string) bool {
	flagPresent := false
	envPresent := false
	if flag := flag.Lookup(flagName); flag != nil && flag.Value.String() != flag.DefValue {
		flagPresent = true
	}
	if _, ok := os.LookupEnv(envName); ok {
		envPresent = true
	}
	return flagPresent || envPresent
}
func (c *Config) UmarshalJSON(data []byte) (err error) {
	var tmp struct {
		Server          string `json:"address" env:"ADDRESS" envDefault:"localhost:8080"`
		StoreFile       string `json:"store_file" env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
		KeyFileFlag     string `env:"KEY"`
		DatabaseEnv     string `json:"database_dsn" env:"DATABASE_DSN"`
		StoreInterval   string `json:"store_interval" env:"STORE_INTERVAL" envDefault:"300s"`
		RestoreFlag     bool   `json:"restore" env:"RESTORE" envDefault:"true"`
		PrivateKeyFile  string `json:"crypto_key" env:"CRYPTO_KEY"`
		configFile      string `env:"CONFIG"`
		TrustedSubnet   string `json:"trusted_subnet" env:"TRUSTED_SUBNET"`
		CertGRPCFile    string `json:"cert" env:"CERT_FILE" envDefault:""`
		CertKeyGRPCFile string `json:"cert_key" env:"CERT_KEY_FILE" envDefault:""`
	}

	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if !c.isDefault("a", "ADDRESS") {
		c.Server = tmp.Server
	}
	if !c.isDefault("f", "STORE_FILE") {
		c.StoreFile = tmp.StoreFile
	}
	if !c.isDefault("d", "DATABASE_DSN") {
		c.DatabaseEnv = tmp.DatabaseEnv
	}
	if !c.isDefault("i", "STORE_INTERVAL") {
		c.StoreInterval, err = time.ParseDuration(tmp.StoreInterval)
	}
	if !c.isDefault("crypto-key", "CRYPTO_KEY") {
		c.PrivateKeyFile = tmp.PrivateKeyFile
	}
	if !c.isDefault("cert", "CERT_FILE") {
		c.CertGRPCFile = tmp.CertGRPCFile
	}
	if !c.isDefault("cert-key", "CERT_KEY_FILE") {
		c.CertKeyGRPCFile = tmp.CertKeyGRPCFile
	}
	if !c.isDefault("r", "RESTORE") {
		c.RestoreFlag = tmp.RestoreFlag
	}
	return err
}

func NewConfig() *Config {
	flagCfg := parseFlag()
	envCfg := &Config{}
	jsonCfg := &Config{}
	err := env.Parse(envCfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("envCfg:")
	log.Println(envCfg)
	log.Println("flagCfg:")
	log.Println(flagCfg.Server)
	if envCfg.ConfigFile != "" {
		jsonData, err := ioutil.ReadFile(envCfg.ConfigFile)
		if err != nil {
			log.Println(err)
		}
		err = jsonCfg.UmarshalJSON(jsonData)
		if err != nil {
			log.Println(err)
		}
	}
	if flagCfg.ConfigFile != "" {
		jsonData, err := ioutil.ReadFile(flagCfg.ConfigFile)
		if err != nil {
			log.Println(err)
		}
		err = jsonCfg.UmarshalJSON(jsonData)
		if err != nil {
			log.Println(err)
		}
	}
	cfg := &Config{}
	cfg.Server = cfg.coalesceString(envCfg.Server, flagCfg.Server, jsonCfg.Server, "localhost:8080")
	cfg.StoreFile = cfg.coalesceString(envCfg.StoreFile, flagCfg.StoreFile, jsonCfg.StoreFile, "")
	cfg.KeyFileFlag = cfg.coalesceString(envCfg.KeyFileFlag, flagCfg.KeyFileFlag, jsonCfg.KeyFileFlag, "")
	//Пустая для наглядности, я понимаю, что функция сама подставит
	cfg.DatabaseEnv = cfg.coalesceString(envCfg.DatabaseEnv, flagCfg.DatabaseEnv, jsonCfg.DatabaseEnv, "")
	cfg.StoreInterval = cfg.coalesceTime(*jsonCfg)
	cfg.RestoreFlag = cfg.coalesceBool(*jsonCfg)
	cfg.PrivateKeyFile = cfg.coalesceString(envCfg.PrivateKeyFile, flagCfg.PrivateKeyFile, jsonCfg.PrivateKeyFile, "")
	cfg.ConfigFile = cfg.coalesceString(envCfg.ConfigFile, flagCfg.ConfigFile, jsonCfg.ConfigFile, "")
	cfg.TrustedSubnet = cfg.coalesceString(envCfg.TrustedSubnet, flagCfg.TrustedSubnet, jsonCfg.TrustedSubnet, "")
	cfg.CertGRPCFile = cfg.coalesceString(envCfg.CertGRPCFile, flagCfg.CertGRPCFile, jsonCfg.CertGRPCFile, "")
	cfg.CertKeyGRPCFile = cfg.coalesceString(envCfg.CertKeyGRPCFile, flagCfg.CertKeyGRPCFile, jsonCfg.CertKeyGRPCFile, "")
	log.Println("resultCfg")
	log.Println(cfg)
	return cfg
}

func (c Config) coalesceBool(json Config) bool {
	def := false
	if _, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		return ok
	}
	if flagVar := flag.Lookup("i"); flagVar != nil {
		return true
	}
	if !c.isDefault("i", "STORE_INTERVAL") {
		if c.ConfigFile != "" {
			return json.RestoreFlag
		}
	}
	return def
}

func (c Config) coalesceTime(json Config) time.Duration {
	def, err := time.ParseDuration("300s")
	if err != nil {
		log.Println("If you are here then the space-time continuum is destroyed")
	}
	if enVar, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		data, err := time.ParseDuration(enVar)
		if err != nil {
			log.Println("env store interbal var parsing error. setting defualt")
			return def
		}
		return data
	}
	if flagVar := flag.Lookup("i"); flagVar != nil && flagVar.Value.String() != flagVar.DefValue {
		data, err := time.ParseDuration(flagVar.Value.String())
		if err != nil {
			log.Println("flag store interbal var parsing error. setting defualt")
			return def
		}
		return data
	}
	if !c.isDefault("i", "STORE_INTERVAL") {
		if c.ConfigFile != "" {
			return json.StoreInterval
		}
	}
	return def
}

func (c Config) coalesceString(strs ...string) string {
	for _, str := range strs {
		if str != "" {
			return str
		}
	}
	return ""
}
