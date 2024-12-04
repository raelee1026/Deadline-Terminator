package gmail

import (
	"context"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var oauth2Config *oauth2.Config

func InitGmailClient() {
	b, err := os.ReadFile("config/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/gmail.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	oauth2Config = config
}

func GetOAuth2Config() *oauth2.Config {
	return oauth2Config
}

func GetClient(token *oauth2.Token) (*http.Client, error) {
	if oauth2Config == nil {
		InitGmailClient()
	}
	return oauth2Config.Client(context.Background(), token), nil
}
