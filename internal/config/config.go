package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string   `yaml:"env" env-default:"local"`
	GRpc     GRPC     `yaml:"grpc" env-required:"true"`
	Storage  Storage  `yaml:"storage" env-required:"true"`
	JWT      JWT      `yaml:"jwt" env-required:"true"`
	Minio    Minio    `yaml:"minio" env-required:"true"`
	Redis    Redis    `yaml:"redis" env-required:"true"`
	RabbitMq RabbitMq `yaml:"rabbitmq" env-required:"true"`
}

type Minio struct {
	Endpoint  string `yaml:"endpoint" env-required:"true"`
	AccessKey string `yaml:"access_key" env-required:"true"`
	SecretKey string `yaml:"secret_key" env-required:"true"`
	Bucket    string `yaml:"bucket" env-required:"true"`
}

type GRPC struct {
	Port    int           `yaml:"port" env-default:"4041"`
	Timeout time.Duration `yaml:"timeout" env-default:"4s"`
}

type JWT struct {
	TokenTTL    time.Duration `yaml:"token_ttl" env-default:"1h"`
	TokenSecret string        `yaml:"token_secret" env-default:"secret"`
}

type Redis struct {
	Addr     string        `yaml:"addr" env-required:"true"`
	Password string        `yaml:"password" env-default:""`
	DB       int           `yaml:"db" env-default:"0"`
	CacheTTL time.Duration `yaml:"cache_ttl" env-default:"1h"`
}

type RabbitMq struct {
	Addr string `yaml:"addr" env-required:"true"`
}

type Storage struct {
	ConnectionString string `yaml:"connection_string" env-required:"true"`
	MigrationPath    string `yaml:"migration_path" env-required:"true"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
