package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI          string
	MongoDB           string
	ServerPort        string
	JWTSecret         string
	FirebaseCredFile  string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Please not there is no envionment variables available")
	}

	config := &Config{
		MongoURI:         os.Getenv("MONGO_URI"),
		MongoDB:          os.Getenv("MONGO_DB"),
		ServerPort:       os.Getenv("SERVER_PORT"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		FirebaseCredFile: os.Getenv("FIREBASE_CREDENTIALS"),
	}

	return config
}
