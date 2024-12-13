package gemini

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// TestGeminiAPI 调用 Gemini API 并生成内容
func ProcessedTask() {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START json_no_schema]
	model := client.GenerativeModel("gemini-1.5-pro-latest")
	// Ask the model to respond with JSON.
	model.ResponseMIMEType = "application/json"
	prompt := `List a few popular cookie recipes using this JSON schema:
                   Recipe = {'recipeName': string}
	           Return: Array<Recipe>`
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}

	for _, candidate := range resp.Candidates {
		fmt.Println("生成的内容:", candidate.Content)
	}
}

/*package gemini

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "time"
)

type Task struct {
    ID          int    `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Deleted     bool   `json:"deleted"`
}

type ProcessedTask struct {
    ID        int        `json:"id"`
    Title     string     `json:"title"`
    Deadline  *time.Time `json:"deadline,omitempty"`
    Summary   string     `json:"summary"`
    IsDeleted bool       `json:"is_deleted"`
}

type GeminiAPIRequest struct {
    Prompt string `json:"prompt"`
}

type GeminiAPIResponse struct {
    Candidates []struct {
        Content string `json:"content"`
    } `json:"candidates"`
}


func ProcessTasks(inputFile, outputFile string) error {
    data, err := ioutil.ReadFile(inputFile)
    if err != nil {
        return err
    }
    var tasks []Task
    if err := json.Unmarshal(data, &tasks); err != nil {
        return err
    }

    var processedTasks []ProcessedTask
    for _, task := range tasks {
        newDeadline := time.Now().Add(5 * 24 * time.Hour)
        newTitle, newSummary, err := generateContent(task.Title, task.Description, newDeadline)
        if err != nil {
            fmt.Printf("處理任務 %d 時出錯: %v\n", task.ID, err)
            continue
        }

        processedTask := ProcessedTask{
            ID:        task.ID,
            Title:     newTitle,
            Deadline:  &newDeadline,
            Summary:   newSummary,
            IsDeleted: task.Deleted,
        }
        processedTasks = append(processedTasks, processedTask)
    }

    outputData, err := json.MarshalIndent(processedTasks, "", "  ")
    if err != nil {
        return err
    }
    if err := ioutil.WriteFile(outputFile, outputData, 0644); err != nil {
        return err
    }

    return nil
}

func generateContent(title, description string, deadline time.Time) (string, string, error) {
	ctx := context.Background()
	prompt := fmt.Sprintf("根據以下資訊生成新的標題和摘要：\n標題: %s\n描述: %s\n截止日期: %s", title, description, deadline.Format("2006-01-02"))
	response, err := callGeminiAPI(ctx, prompt)
	if err != nil {
			return "", "", fmt.Errorf("呼叫 Gemini API 時出錯: %v", err)
	}

	var apiResponse GeminiAPIResponse
	if err := json.Unmarshal([]byte(response), &apiResponse); err != nil {
			return "", "", fmt.Errorf("解析 Gemini API 回應時出錯: %v", err)
	}

	if len(apiResponse.Candidates) == 0 {
			return "", "", fmt.Errorf("Gemini API 未返回任何候選項")
	}

	generatedContent := apiResponse.Candidates[0].Content
	return generatedContent, generatedContent, nil
}

func callGeminiAPI(ctx context.Context, prompt string) (string, error) {
	model := "gemini-1.5-flash" // 請替換為您使用的模型名稱
	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta2/models/%s:generateContent", model)

	requestBody, err := json.Marshal(GeminiAPIRequest{Prompt: prompt})
	if err != nil {
			return "", fmt.Errorf("構建請求體時出錯: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
			return "", fmt.Errorf("創建 HTTP 請求時出錯: %v", err)
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
			return "", fmt.Errorf("未設置 API_KEY 環境變數")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
			return "", fmt.Errorf("發送 HTTP 請求時出錯: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("API 請求失敗，狀態碼: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
			return "", fmt.Errorf("讀取回應內容時出錯: %v", err)
	}

	if len(body) == 0 {
			return "", fmt.Errorf("API 回應內容為空")
	}

	return string(body), nil
}*/
