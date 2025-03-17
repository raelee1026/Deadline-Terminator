package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"encoding/json"
	"io/ioutil"
	"os"
)

type OAuthCredentials struct {
	Web struct {
		ClientID     string   `json:"client_id"`
		ProjectID    string   `json:"project_id"`
		AuthURI      string   `json:"auth_uri"`
		TokenURI     string   `json:"token_uri"`
		CertURL      string   `json:"auth_provider_x509_cert_url"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
	} `json:"web"`
}

func loadCredentials(filename string) (*OAuthCredentials, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open credentials file: %v", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %v", err)
	}

	var credentials OAuthCredentials
	if err := json.Unmarshal(data, &credentials); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %v", err)
	}

	return &credentials, nil
}

func main() {
	credentials, err := loadCredentials("config/credentials.json")
	if err != nil {
		log.Fatalf("Error loading credentials: %v", err)
	}

	// Construct authorization URL dynamically
	authURL := fmt.Sprintf("%s?access_type=offline&client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=state-token",
		credentials.Web.AuthURI,
		credentials.Web.ClientID,
		credentials.Web.RedirectURIs[0],
		"https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fgmail.readonly",
	)

	fmt.Printf("Please visit the following URL to complete the authorization:\n%s\n", authURL)
	StartServer()
}

/*func main() {
	// go openBrowser("https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=997285622302-goltvajj196rm1ims0sijhgbvro82cad.apps.googleusercontent.com&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Foauth2%2Fcallback&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fgmail.readonly&state=state-token")
	url := "https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=997285622302-goltvajj196rm1ims0sijhgbvro82cad.apps.googleusercontent.com&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Foauth2%2Fcallback&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fgmail.readonly&state=state-token"
	fmt.Printf("Please visit the following URL to complete the authorization:\n%s\n", url)
	StartServer()
}*/

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Fatalf("Failed to open browser: %v", err)
	}
}
