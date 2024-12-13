package gemini

// AccountInfo 定義獲取賬戶資訊的回應結構
type AccountInfo struct {
	Username string `json:"username"`
	Balance  string `json:"balance"`
	// 根據 API 文件添加更多欄位
}
