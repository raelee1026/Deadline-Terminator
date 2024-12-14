package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var (
	oauth2Config *oauth2.Config
	token        *oauth2.Token
)

var messages []gmail.Message

func init() {
	b, err := os.ReadFile("../config/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	oauth2Config = config
}

func HandleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}

	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := oauth2Config.Client(context.Background(), token)
	srv, err := gmail.New(client)
	if err != nil {
		http.Error(w, "Unable to retrieve Gmail client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var errMsg error
	messages, errMsg := GetFilteredMessages(srv)
	if errMsg != nil {
		http.Error(w, "Unable to retrieve messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	getMessages(messages)

	w.Header().Set("Content-Type", "application/json")
	http.Redirect(w, r, "http://localhost:8080", http.StatusSeeOther)
	json.NewEncoder(w).Encode(messages)
}

func GetFilteredMessages(srv *gmail.Service) ([]gmail.Message, error) {
	user := "me"
	req := srv.Users.Messages.List(user).LabelIds("INBOX").MaxResults(50)
	res, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages: %v", err)
	}

	var messages []gmail.Message
	for _, m := range res.Messages {
		msg, err := srv.Users.Messages.Get(user, m.Id).Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve message: %v", err)
		}

		for _, header := range msg.Payload.Headers {
			if header.Name == "Subject" && strings.HasPrefix(header.Value, "1131.") {
				messages = append(messages, *msg)
			}
		}
	}
	return messages, nil
}

func getGmailClient() (*http.Client, error) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok), nil
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser and authorize the application: \n%v\n", authURL)

	var authCode string
	fmt.Print("Enter the authorization code: ")
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
