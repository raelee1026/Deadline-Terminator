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

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var tasks []Task
var nextID = 1

func init() {
	err := loadTasksFromFile("../backend/jsonfortest/tasks.json")
	if err != nil {
		log.Fatalf("Failed to load tasks: %v", err)
	}
}

func StartServer() {
	http.HandleFunc("/api/tasks", handleTasks)
	http.HandleFunc("/api/tasks/delete", handleDeleteTask)
	http.HandleFunc("/api/tasks/sync", handleSyncTasks)
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
			tasks[i].Deleted = true
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

	tasks = append(tasks, newTasks...)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newTasks)
}

func SortTasksByDeadline() {
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Deleted != tasks[j].Deleted {
			return !tasks[i].Deleted
		}
		return tasks[i].Deadline.Before(tasks[j].Deadline)
	})
}

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

	for _, task := range tasks {
		if task.ID >= nextID {
			nextID = task.ID + 1
		}
	}

	log.Printf("Loaded %d tasks from file", len(tasks))
	return nil
}

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
