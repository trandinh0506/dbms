package pkg

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL     string
	JWTSecret []byte
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	return &Config{
		DBURL:     os.Getenv("DB_URL"),
		JWTSecret: []byte([]byte(os.Getenv("JWTSecret"))),
	}

}
