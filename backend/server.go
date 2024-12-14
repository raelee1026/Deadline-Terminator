package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var tasks [][]Task
var nextID = 1
var message []gmail.Message

func init() {
	// 初始化 tasks 切片
	tasks = make([][]Task, 2)

	// 加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Error loading .env file")
	}

	// 加载任务文件
	filenames := []string{"../backend/Task/tasks.json", "../backend/Task/rowTasks.json"}
	if err := loadTasksFromFile(filenames); err != nil {
		log.Fatalf("Failed to load tasks: %v", err)
	}
}

func StartServer() {
	http.HandleFunc("/api/tasks", handleTasks)
	http.HandleFunc("/api/tasks/delete", handleDeleteTask)
	http.HandleFunc("/api/tasks/sync", handleSyncTasks)
	http.HandleFunc("/api/tasks/catch", handleCatchMessages)
	http.HandleFunc("/oauth2/callback", HandleOAuth2Callback)
	http.Handle("/", http.FileServer(http.Dir("../frontend")))

	log.Println("Starting Deadline Terminator server on :8080")
	log.Println("Visit the authentication URL to authenticate with Gmail")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Deadline    time.Time `json:"deadline"`
	Description string    `json:"description"`
	Deleted     bool      `json:"deleted"`
}

func getMessages(content []gmail.Message) {
	message = content
}

func handleCatchMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ProcessMessages(message)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Gmail synced successfully",
	})
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		SortTasksByDeadline()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks[0])
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
		tasks[0] = append(tasks[0], task)

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

	for i, task := range tasks[0] {
		if task.ID == request.ID {
			tasks[0][i].Deleted = true
			tasks[1][i].Deleted = true
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Task marked as deleted")
			return
		}
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}

func handleSyncTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	newTasks, err := syncGmailTasks()
	if err != nil {
		http.Error(w, "Failed to sync Gmail tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tasks[0] = append(tasks[0], newTasks...)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newTasks)
}

func SortTasksByDeadline() {
	sort.Slice(tasks[0], func(i, j int) bool {
		if tasks[0][i].Deleted != tasks[0][j].Deleted {
			return !tasks[0][i].Deleted
		}
		return tasks[0][i].Deadline.Before(tasks[0][j].Deadline)
	})
}

func loadTasksFromFile(filenames []string) error {
	for index, filename := range filenames {
		var newNextID = 1
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("Could not open file %s: %v", filename, err)
			continue
		}
		defer file.Close()

		data, err := ioutil.ReadAll(file)
		if err != nil {
			log.Printf("Could not read file %s: %v", filename, err)
			continue
		}

		var fileTasks []Task
		if err := json.Unmarshal(data, &fileTasks); err != nil {
			log.Printf("Could not parse JSON from file %s: %v", filename, err)
			continue
		}

		tasks[index] = append(tasks[index], fileTasks...)
		for _, task := range fileTasks {
			if task.ID >= nextID {
				newNextID = task.ID + 1
			}
		}

		log.Printf("Loaded %d tasks from file %s", len(fileTasks), filename)
		nextID = max(newNextID, nextID)
	}
	return nil
}

func saveTasksToFile(filename string) error {
	data, err := json.MarshalIndent(tasks[0], "", "  ")
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

func syncGmailTasks() ([]Task, error) {
	client, err := getGmailClient()
	if err != nil {
		return nil, err
	}

	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Gmail client: %v", err)
	}

	messages, err := GetFilteredMessages(srv)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages: %v", err)
	}

	var newTasks []Task
	for _, msg := range messages {
		var subject string
		for _, header := range msg.Payload.Headers {
			if header.Name == "Subject" {
				subject = header.Value
				break
			}
		}

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
