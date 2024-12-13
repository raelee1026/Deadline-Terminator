package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"quickstart/gemini"

	"github.com/joho/godotenv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var (
	oauth2Config *oauth2.Config
	token        *oauth2.Token
)

func init() {
	b, err := os.ReadFile("../config/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// load env
	err = godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	oauth2Config = config
}

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Deadline    time.Time `json:"deadline"`
	Description string    `json:"description"`
	Deleted     bool      `json:"deleted"`
}

var tasks []Task
var nextID = 1

func main() {
	err := loadTasksFromFile("../backend/jsonfortest/tasks.json")
	if err != nil {
		log.Fatalf("Failed to load tasks: %v", err)
	}

	gemini.ProcessedTask()

	http.HandleFunc("/api/tasks", handleTasks)
	http.HandleFunc("/api/tasks/delete", handleDeleteTask)
	http.HandleFunc("/api/tasks/sync", handleSyncTasks)
	http.HandleFunc("/oauth2/callback", handleOAuth2Callback)
	http.Handle("/", http.FileServer(http.Dir("../frontend")))

	log.Println("Starting Deadline Terminator server on :8080")
	log.Println("Visit https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=997285622302-goltvajj196rm1ims0sijhgbvro82cad.apps.googleusercontent.com&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Foauth2%2Fcallback&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fgmail.readonly&state=state-token to authenticate with Gmail")
	// I want to use credentials.json to replace the client_id and client_secret in the URL
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// avoid repeated task in task.json
func isTaskExists(subject string) bool {
	for _, task := range tasks {
		if task.Title == subject {
			return true
		}
	}
	return false
}

func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
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

	user := "me"
	req := srv.Users.Messages.List(user).LabelIds("INBOX").MaxResults(50)
	res, err := req.Do()
	if err != nil {
		http.Error(w, "Unable to retrieve messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var newTasks []Task

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<html><body><h1>OAuth2 Callback Page</h1>"))
	w.Write([]byte("<h2>Filtered Inbox Data:</h2><ul>"))

	for _, m := range res.Messages {
		msg, err := srv.Users.Messages.Get(user, m.Id).Do()
		if err != nil {
			http.Error(w, "Unable to retrieve message: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var subject, body string
		for _, header := range msg.Payload.Headers {
			if header.Name == "Subject" && strings.HasPrefix(header.Value, "1131.") {
				subject = header.Value
			}
		}

		if subject == "" || isTaskExists(subject) {
			// Skip if subject is empty or task already exists
			continue
		}

		for _, part := range msg.Payload.Parts {
			if part.MimeType == "text/plain" {
				decodedBody, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err != nil {
					body = "Unable to decode body"
				} else {
					body = string(decodedBody)
				}
				break
			}
		}

		// create task
		newTask := Task{
			ID:          nextID,
			Title:       subject,
			Deadline:    time.Now().AddDate(0, 0, 7), // assumption
			Description: body,
			Deleted:     false,
		}
		nextID++
		newTasks = append(newTasks, newTask)

		w.Write([]byte("<li><h3>" + subject + "</h3>" + body + "</li>"))
	}

	// save to task
	tasks = append(tasks, newTasks...)
	err = saveTasksToFile("../backend/jsonfortest/tasks.json")
	if err != nil {
		http.Error(w, "Failed to save tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("</ul></body></html>"))
}

// handleTasks 處理任務的 GET 和 POST 請求
func handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		SortTasksByDeadline()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
		return
	} else if r.Method == http.MethodPost {
		var task Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, "Invalid task format", http.StatusBadRequest)
			return
		}

		task.ID = nextID
		nextID++
		tasks = append(tasks, task)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
		return
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ID int `json:"id"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		fmt.Println("Error decoding request:", err)
		return
	}

	for i, task := range tasks {
		if task.ID == request.ID {
			// 標記為刪除
			tasks[i].Deleted = true
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Task marked as deleted")
			return
		}
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}

// handleSyncTasks 同步 Gmail 收件匣中的郵件作為任務
func handleSyncTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 呼叫 Gmail API 並同步郵件為任務
	newTasks, err := syncGmailTasks()
	if err != nil {
		http.Error(w, "Failed to sync Gmail tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 合併到任務列表
	tasks = append(tasks, newTasks...)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newTasks)
}

// SortTasksByDeadline 將任務按截止日期排序
func SortTasksByDeadline() {
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Deleted != tasks[j].Deleted {
			return !tasks[i].Deleted
		}
		return tasks[i].Deadline.Before(tasks[j].Deadline)
	})
}

// syncGmailTasks 從 Gmail 收件匣中同步郵件作為任務
func syncGmailTasks() ([]Task, error) {
	client, err := getGmailClient()
	if err != nil {
		return nil, err
	}

	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Gmail client: %v", err)
	}

	// 列出收件匣中的郵件
	user := "me"
	messages, err := srv.Users.Messages.List(user).LabelIds("INBOX").MaxResults(10).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages: %v", err)
	}

	var newTasks []Task
	for _, msg := range messages.Messages {
		fullMessage, err := srv.Users.Messages.Get(user, msg.Id).Format("full").Do()
		if err != nil {
			log.Printf("Unable to retrieve message %s: %v", msg.Id, err)
			continue
		}

		// 提取郵件標題
		var subject string
		for _, header := range fullMessage.Payload.Headers {
			if header.Name == "Subject" {
				subject = header.Value
				break
			}
		}

		// 創建任務
		newTask := Task{
			ID:          nextID,
			Title:       subject,
			Deadline:    time.Now().AddDate(0, 0, 7), // 假設截止日期為一周後
			Description: "Imported from Gmail",
			Deleted:     false,
		}
		nextID++
		newTasks = append(newTasks, newTask)
	}

	return newTasks, nil
}

// Gmail OAuth 驗證流程
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

// https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=997285622302-goltvajj196rm1ims0sijhgbvro82cad.apps.googleusercontent.com&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Foauth2%2Fcallback&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fgmail.readonly&state=state-token

// loadTasksFromFile 讀取 JSON 文件並初始化任務清單
func loadTasksFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("could not read file: %v", err)
	}

	err = json.Unmarshal(data, &tasks)
	if err != nil {
		return fmt.Errorf("could not parse JSON: %v", err)
	}

	// 設置 nextID 為當前最大 ID + 1
	for _, task := range tasks {
		if task.ID >= nextID {
			nextID = task.ID + 1
		}
	}

	log.Printf("Loaded %d tasks from file", len(tasks))
	return nil
}

// saveTasksToFile 將任務清單保存到 JSON 文件
func saveTasksToFile(filename string) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal tasks: %v", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("could not write file: %v", err)
	}

	log.Println("Tasks saved to file")
	return nil
}
