package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
)

// Task represents a task or exam
type Task struct {
	ID          int       `json:"id"`          // 唯一識別符
	Title       string    `json:"title"`       // 任務標題
	Deadline    time.Time `json:"deadline"`    // 截止日期
	Description string    `json:"description"` // 任務描述
	Deleted     bool      `json:"deleted"`
}

var tasks []Task
var nextID = 1 // 自增量，用於生成唯一的任務 ID

func main() {
	http.HandleFunc("/api/tasks", handleTasks)
	http.HandleFunc("/api/tasks/delete", handleDeleteTask)

	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Starting Deadline Terminator server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// SortTasksByDeadline sorts tasks by their deadline
func SortTasksByDeadline() {
	sort.Slice(tasks, func(i, j int) bool {
		// 如果一個任務已刪除，則它永遠排在後面
		if tasks[i].Deleted != tasks[j].Deleted {
			return !tasks[i].Deleted
		}
		// 如果兩個任務都未刪除或都已刪除，按截止日期排序
		return tasks[i].Deadline.After(tasks[j].Deadline)
	})
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		SortTasksByDeadline()
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

		// 分配唯一 ID
		task.ID = nextID
		nextID++

		tasks = append(tasks, task)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task) // 返回新增的任務
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
