package gmail

import (
	"Deadline-Terminator/backend/tasks"
	"context"
	"log"
	"net/http"
	"time"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func HandleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}

	token, err := GetOAuth2Config().Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	SaveToken("token.json", token)
	http.Redirect(w, r, "/", http.StatusFound)
}

func HandleSyncTasks(w http.ResponseWriter, r *http.Request) {
	token, err := TokenFromFile("token.json")
	if err != nil {
		http.Error(w, "Unable to read token file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	client, err := GetClient(token)
	if err != nil {
		http.Error(w, "Unable to create Gmail client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		http.Error(w, "Unable to retrieve Gmail service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user := "me"
	messages, err := srv.Users.Messages.List(user).LabelIds("INBOX").MaxResults(50).Do()
	if err != nil {
		http.Error(w, "Unable to retrieve messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, msg := range messages.Messages {
		fullMsg, err := srv.Users.Messages.Get(user, msg.Id).Format("full").Do()
		if err != nil {
			log.Printf("Unable to retrieve message %s: %v", msg.Id, err)
			continue
		}
		var subject string
		for _, header := range fullMsg.Payload.Headers {
			if header.Name == "Subject" {
				subject = header.Value
				break
			}
		}
		tasks.Task{
			ID:          tasks.NextID(),
			Title:       subject,
			Deadline:    time.Now().AddDate(0, 0, 7),
			Description: "Imported from Gmail",
		}
	}
}
