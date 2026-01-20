package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type (
	Server struct {
		Mode         string `env:"MODE"`
		Port         string `env:"PORT"`
		JWTSecretKey string `env:"JWT_SECRET_KEY"`
	}

	Client struct {
		Url  string `env:"FRONTEND_URL"`
		Port string `env:"CLIENT_PORT"`
	}

	WalletService struct {
		BaseURL string `env:"WALLET_SERVICE_URL"`
	}

	Database struct {
		DBHost     string `env:"DB_HOST"`
		DBPort     string `env:"DB_PORT"`
		DBUser     string `env:"DB_USER"`
		DBPassword string `env:"DB_PASSWORD"`
		DBName     string `env:"DB_NAME"`
	}

	Minio struct {
		Host        string `env:"MINIO_HOST"`
		AccessKey   string `env:"MINIO_ROOT_USER"`
		SecretKey   string `env:"MINIO_ROOT_PASSWORD"`
		MaxOpenConn int    `env:"MINIO_MAX_OPEN_CONN"`
		UseSSL      int    `env:"MINIO_USE_SSL"`
	}

	Config struct {
		Server        Server
		Client        Client
		WalletService WalletService
		Database      Database
		Minio         Minio
	}
)

var Cfg Config

func LoadNative() ([]string, error) {
	var ok bool
	var missing []string

	if _, err := os.Stat("/app/.env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, err
		}
	}

	// ! Load Server configuration ____________________________
	if Cfg.Server.Mode, ok = os.LookupEnv("MODE"); !ok {
		missing = append(missing, "MODE env is not set")
	}
	if Cfg.Server.Port, ok = os.LookupEnv("PORT"); !ok {
		missing = append(missing, "PORT env is not set")
	}
	if Cfg.Server.JWTSecretKey, ok = os.LookupEnv("JWT_SECRET_KEY"); !ok {
		missing = append(missing, "JWT_SECRET_KEY env is not set")
	}
	// ! ______________________________________________________

	// ! Load Client configuration ____________________________
	if Cfg.Client.Url, ok = os.LookupEnv("FRONTEND_URL"); !ok {
		missing = append(missing, "FRONTEND_URL env is not set")
	}
	if Cfg.Client.Port, ok = os.LookupEnv("CLIENT_PORT"); !ok {
		missing = append(missing, "CLIENT_PORT env is not set")
	}
	// ! ______________________________________________________

	// ! Load Wallet Service configuration ___________________
	if Cfg.WalletService.BaseURL, ok = os.LookupEnv("WALLET_SERVICE_URL"); !ok {
		missing = append(missing, "WALLET_SERVICE_URL env is not set")
	}
	// ! ______________________________________________________

	// ! Load Database configuration __________________________
	if Cfg.Database.DBUser, ok = os.LookupEnv("DB_USER"); !ok {
		missing = append(missing, "DB_USER env is not set")
	}
	if Cfg.Database.DBHost, ok = os.LookupEnv("DB_HOST"); !ok {
		missing = append(missing, "DB_HOST env is not set")
	}
	if Cfg.Database.DBPort, ok = os.LookupEnv("DB_PORT"); !ok {
		missing = append(missing, "DB_PORT env is not set")
	}
	if Cfg.Database.DBName, ok = os.LookupEnv("DB_NAME"); !ok {
		missing = append(missing, "DB_NAME env is not set")
	}
	if Cfg.Database.DBPassword, ok = os.LookupEnv("DB_PASSWORD"); !ok {
		missing = append(missing, "DB_PASSWORD env is not set")
	}
	// ! ______________________________________________________

	// ! Load MinIO configuration _____________________________
	if Cfg.Minio.Host, ok = os.LookupEnv("MINIO_HOST"); !ok {
		missing = append(missing, "MINIO_HOST env is not set")
	}
	if Cfg.Minio.AccessKey, ok = os.LookupEnv("MINIO_ROOT_USER"); !ok {
		missing = append(missing, "MINIO_ROOT_USER env is not set")
	}
	if Cfg.Minio.SecretKey, ok = os.LookupEnv("MINIO_ROOT_PASSWORD"); !ok {
		missing = append(missing, "MINIO_ROOT_PASSWORD env is not set")
	}
	if val, ok := os.LookupEnv("MINIO_MAX_OPEN_CONN"); !ok {
		missing = append(missing, "MINIO_MAX_OPEN_CONN env is not set")
	} else {
		var err error
		if Cfg.Minio.MaxOpenConn, err = strconv.Atoi(val); err != nil {
			missing = append(missing, fmt.Sprintf("MINIO_MAX_OPEN_CONN must be int, got %s", val))
		}
	}
	if val, ok := os.LookupEnv("MINIO_USE_SSL"); !ok {
		missing = append(missing, "MINIO_USE_SSL env is not set")
	} else {
		var err error
		if Cfg.Minio.UseSSL, err = strconv.Atoi(val); err != nil {
			missing = append(missing, fmt.Sprintf("MINIO_USE_SSL must be int, got %s", val))
		}
	}
	// ! ______________________________________________________

	return missing, nil
}

func LoadByViper() ([]string, error) {
	var missing []string

	config := viper.New()
	if configFile, err := os.Stat("/app/config.json"); err != nil || configFile.IsDir() {
		config.SetConfigFile("config.json")
	} else {
		config.SetConfigFile("/app/config.json")
	}

	if err := config.ReadInConfig(); err != nil {
		return nil, err
	}

	// ! Load Server configuration ____________________________
	if Cfg.Server.Mode = config.GetString("MODE"); Cfg.Server.Mode == "" {
		missing = append(missing, "MODE env is not set")
	}
	if Cfg.Server.Port = config.GetString("PORT"); Cfg.Server.Port == "" {
		missing = append(missing, "PORT env is not set")
	}
	if Cfg.Server.JWTSecretKey = config.GetString("JWT_SECRET_KEY"); Cfg.Server.JWTSecretKey == "" {
		missing = append(missing, "JWT_SECRET_KEY env is not set")
	}
	// ! ______________________________________________________

	// ! Load Client configuration ____________________________
	if Cfg.Client.Url = config.GetString("CLIENT.URL"); Cfg.Client.Url == "" {
		missing = append(missing, "CLIENT.URL env is not set")
	}
	if Cfg.Client.Port = config.GetString("CLIENT.PORT"); Cfg.Client.Port == "" {
		missing = append(missing, "CLIENT.PORT env is not set")
	}
	// ! ______________________________________________________

	// ! Load Wallet Service configuration ___________________
	if Cfg.WalletService.BaseURL = config.GetString("WALLET_SERVICE.URL"); Cfg.WalletService.BaseURL == "" {
		missing = append(missing, "WALLET_SERVICE.URL env is not set")
	}
	// ! ______________________________________________________

	// ! Load Database configuration __________________________
	if Cfg.Database.DBUser = config.GetString("DATABASE.POSTGRESQL.USER"); Cfg.Database.DBUser == "" {
		missing = append(missing, "DATABASE.POSTGRESQL.USER env is not set")
	}
	if Cfg.Database.DBHost = config.GetString("DATABASE.POSTGRESQL.HOST"); Cfg.Database.DBHost == "" {
		missing = append(missing, "DATABASE.POSTGRESQL.HOST env is not set")
	}
	if Cfg.Database.DBPort = config.GetString("DATABASE.POSTGRESQL.PORT"); Cfg.Database.DBPort == "" {
		missing = append(missing, "DATABASE.POSTGRESQL.PORT env is not set")
	}
	if Cfg.Database.DBName = config.GetString("DATABASE.POSTGRESQL.NAME"); Cfg.Database.DBName == "" {
		missing = append(missing, "DATABASE.POSTGRESQL.NAME env is not set")
	}
	if Cfg.Database.DBPassword = config.GetString("DATABASE.POSTGRESQL.PASSWORD"); Cfg.Database.DBPassword == "" {
		missing = append(missing, "DATABASE.POSTGRESQL.PASSWORD env is not set")
	}
	// ! ______________________________________________________

	// ! Load Minio configuration __________________________
	if Cfg.Minio.Host = config.GetString("OBJECT-STORAGE.MINIO.HOST"); Cfg.Minio.Host == "" {
		missing = append(missing, "OBJECT-STORAGE.MINIO.HOST env is not set")
	}
	if Cfg.Minio.AccessKey = config.GetString("OBJECT-STORAGE.MINIO.USER"); Cfg.Minio.AccessKey == "" {
		missing = append(missing, "OBJECT-STORAGE.MINIO.USER env is not set")
	}
	if Cfg.Minio.SecretKey = config.GetString("OBJECT-STORAGE.MINIO.PASSWORD"); Cfg.Minio.SecretKey == "" {
		missing = append(missing, "OBJECT-STORAGE.MINIO.PASSWORD env is not set")
	}
	if Cfg.Minio.MaxOpenConn = config.GetInt("OBJECT-STORAGE.MINIO.MAX_OPEN_CONN_POOL"); Cfg.Minio.MaxOpenConn == 0 {
		missing = append(missing, "OBJECT-STORAGE.MINIO.MAX_OPEN_CONN_POOL env is not set")
	}
	if Cfg.Minio.UseSSL = config.GetInt("OBJECT-STORAGE.MINIO.USE_SSL"); Cfg.Minio.UseSSL < 0 || Cfg.Minio.UseSSL > 1 {
		missing = append(missing, "OBJECT-STORAGE.MINIO.USE_SSL env is not valid")
	}

	// ! ______________________________________________________

	return missing, nil
}
