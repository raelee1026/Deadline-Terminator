package gmail

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
)

func TokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func SaveToken(file string, token *oauth2.Token) {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Unable to save token to file: %v", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
