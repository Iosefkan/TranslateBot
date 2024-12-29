package environ

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Environments struct {
	BotToken string
}

func GetEnvironments() Environments {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		fmt.Printf("File does not exist\n")
	} else {
		err := godotenv.Load(".env")
		if err != nil {
			return Environments{}
		}
	}

	botToken := os.Getenv("BOT_TOKEN")
	return Environments{BotToken: botToken}
}
