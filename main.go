package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Task struct {
	Title       string    `json:"title"`
	Deadline    time.Time `json:"deadline"`
	Description string    `json:"description"`
}

var tasks []Task

func main() {
	http.HandleFunc("/api/tasks", handleTasks)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Starting Deadline Terminator server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		if tasks == nil {
			tasks = []Task{}
		}
		json.NewEncoder(w).Encode(tasks)
		return
	} else if r.Method == http.MethodPost {
		var task Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, "Invalid task format", http.StatusBadRequest)
			fmt.Println("Error decoding task:", err) // 調試用輸出
			return
		}
		tasks = append(tasks, task)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task) // 返回新增的任務
		return
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
