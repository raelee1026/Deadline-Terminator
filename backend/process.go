package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	"strings"
	"regexp"

	"google.golang.org/api/gmail/v1"
)

type Content struct {
	Parts []string `json:Parts`
	Role  string   `json:Role`
}

var CourseNames []string
//var CourseCount = 0

// avoid repeated task in task.json
func isTaskExists(subject string) bool {
	if tasks == nil || len(tasks[1]) == 0 {
		return false
	}
	for _, task := range tasks[1] {
		if task.Title == subject && !task.Deleted {
			return true
		}
	}
	return false
}

// ProcessMessages processes all Gmail messages in a single batch for Gemini
func ProcessMessages(messages []gmail.Message) {
	filterPrefix := os.Getenv("FILTER_PREFIX")
	if len(messages) == 0 {
		log.Println("No messages to process.")
		return
	}

	// fmt.Println(messages)
	for _, msg := range messages {
		var subject, body string

		// Extract subject
		for _, header := range msg.Payload.Headers {
			if header.Name == "Subject" {
				subject = header.Value
				break
			}
		}

		// Extract body
		for _, part := range msg.Payload.Parts {
			if part.MimeType == "text/plain" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err != nil {
					log.Printf("Failed to decode body: %v", err)
					continue
				}
				body = string(data)
				break
			}
		}

		if subject == "" || isTaskExists(subject) {
			continue
		}

		// Append Original
		originalTask := Task{
			ID:          nextID,
			Title:       subject,
			Deadline:    time.Now().AddDate(0, 0, 4),
			Description: body,
			Deleted:     false,
		}
		tasks[1] = append(tasks[1], originalTask)

		// Append Tasks

		re := regexp.MustCompile(`^`+ filterPrefix + `\.\d+[:.]?\s*`)
		subject = re.ReplaceAllString(subject, "")

		newSubject := ""
		var courseName string
		if len(CourseNames) > 0 {
			courseName = CourseNames[0]
			CourseNames = CourseNames[1:]

			if !strings.Contains(subject, courseName) {	
				newSubject = fmt.Sprintf("%s %s", courseName, subject)
			} else {
				newSubject = subject
			}
		}	else {
			newSubject = subject
		}

		tasks[0] = append(tasks[0], Task{
			ID:          nextID,
			Title:       newSubject,
			Deadline:    time.Now().AddDate(0, 0, 4),
			Description: body,
			Deleted:     false,
		})
		saveGeneratedTask()
		nextID++
	}
}

// saveGeneratedTask saves the generated JSON to a file
func saveGeneratedTask() {

	// filenames := []string{"Task/tasks.json", "Task/rowTasks.json"}
	// docker
	filenames := []string{"/app/Task/tasks.json", "/app/Task/rowTasks.json"}

	formattedJSON, err := json.MarshalIndent(tasks[0], "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}
	rowData, err := json.MarshalIndent(tasks[1], "", " ")

	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	err = os.WriteFile(filenames[0], formattedJSON, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	err = os.WriteFile(filenames[1], rowData, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("JSON data written to", filenames[0])
}