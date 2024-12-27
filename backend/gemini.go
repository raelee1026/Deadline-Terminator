package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Content struct {
	Parts []string `json:Parts`
	Role  string   `json:Role`
}
type Candidates struct {
	Content *Content `json:Content`
}
type ContentResponse struct {
	Candidates *[]Candidates `json:Candidates`
}

var CourseNames []string

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
	if len(messages) == 0 {
		log.Println("No messages to process.")
		return
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro")
	model.ResponseMIMEType = "application/json"
	// Build a single prompt for all messages

	prompt := ""
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

		// 確保 CourseNames 的索引有效
		/*courseName := ""
		if i < len(CourseNames) {
			courseName = CourseNames[i]
		} else {
			courseName = ""
		}*/

		// 動態生成包含 CourseNames 的 prompt
		taskPrompt := fmt.Sprintf(`
		Generate a task for the email. Use the JSON format provided.
		The title should be concise (less than 20 characters).%s
		Use Traditional Chinese for Chinese emails, English for English emails, and Japanese for Japanese emails.
		The description should be detailed and include line breaks where appropriate for better readability.
		The "id" should be %d.
		The "deadline" should be 4 days after the current date (UTC+08:00).
		The "deleted" field should always be false.

		Email details:
		Subject: %s
		Body: %s

		Output format:
		[
			{
				"id": %d,
				"title": "string",
				"deadline": "ISO 8601 formatted date string",
				"description": "string",
				"deleted": false
			}
		]`, "", nextID, subject, body, nextID)

		// Append this prompt to the main prompt
		prompt += taskPrompt
		nextID++
	}

	// Send the single prompt to Gemini
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatalf("Failed to generate tasks: %v", err)
	}

	marshalResponse, _ := json.MarshalIndent(resp, "", "  ")
	var generateResponse ContentResponse
	if err := json.Unmarshal(marshalResponse, &generateResponse); err != nil {
		log.Fatal(err)
	}

	for _, cad := range *generateResponse.Candidates {
		if cad.Content != nil {
			for _, part := range cad.Content.Parts {
				saveGeneratedTask(part)
			}
		}
	}
}

// saveGeneratedTask saves the generated JSON to a file
func saveGeneratedTask(content string) {

	filenames := []string{"/app/Task/tasks.json", "/app/Task/rowTasks.json"}
	// 解析新生成的內容
	var newTasks []Task
	err := json.Unmarshal([]byte(content), &newTasks)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	for i := range newTasks {
		if i < len(CourseNames) {
			newTasks[i].Title = fmt.Sprintf("%s %s", CourseNames[i], newTasks[i].Title)
		} else {
			newTasks[i].Title = fmt.Sprintf("%s", newTasks[i].Title)
		}
	}

	// 合併 tasks
	tasks[0] = append(tasks[0], newTasks...)

	// 寫入文件
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