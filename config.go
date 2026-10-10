//config.go
//Loads environment specific configuration.
//The active environment is chosen with the APP_ENV variable (dev or prod)
//and the matching .env.<APP_ENV> file is loaded. Real environment variables
//always win because godotenv.Load never overwrites a variable that is set.

package main

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/joho/godotenv"
)

//loadEnv loads the .env file for the current APP_ENV, defaulting to dev.
//A missing file is not an error: in production the variables are usually
//provided directly by the host (systemd, Docker, etc.) with no file at all.
func loadEnv() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	file := ".env." + env
	if err := godotenv.Load(file); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("Warning: could not load %s: %v", file, err)
		}
	}
}

//appID returns the Discord application ID for the current environment.
//It is read from APP_ID so that dev and prod can use different bots.
func appID() discord.AppID {
	raw := os.Getenv("APP_ID")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		log.Fatalf("APP_ID must be set to a valid Discord application ID (got %q)", raw)
	}
	return discord.AppID(id)
}
