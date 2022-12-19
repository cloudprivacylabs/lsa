package cmd

import "github.com/joho/godotenv"

var Environment = make(map[string]string)

func LoadConfig() {
	env, _ := godotenv.Read(".env")
	Environment = env
}
