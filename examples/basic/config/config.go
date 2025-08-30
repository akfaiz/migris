package config

import "github.com/joho/godotenv"

type Config struct {
	Database Database
}

func Load() (Config, error) {
	var config Config
	err := godotenv.Load()
	if err != nil {
		return config, err
	}
	config = Config{
		Database: getDatabaseConfig(),
	}

	return config, nil
}
