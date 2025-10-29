package gotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type GotifyMessage struct {
	Title    string `json:"title,omitempty"`
	Message  string `json:"message"`
	Priority int    `json:"priority,omitempty"`
}

// SendMessage 发送Gotify消息
// serverURL: Gotify服务器地址 (例如: http://your-gotify-server.com 或 https://gotify.example.com)
// token: Gotify应用Token
// title: 消息标题
// message: 消息内容
// priority: 消息优先级 (0-10, 默认为5)
func SendMessage(serverURL, token, title, message string, priority int) error {
	// 确保serverURL不以斜杠结尾
	serverURL = strings.TrimSuffix(serverURL, "/")

	// 构造完整URL
	url := fmt.Sprintf("%s/message?token=%s", serverURL, token)

	msg := GotifyMessage{
		Title:    title,
		Message:  message,
		Priority: priority,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取响应体以获取更多错误信息
		var respBody bytes.Buffer
		_, err := respBody.ReadFrom(resp.Body)
		if err != nil {
			return fmt.Errorf("unexpected status code: %d, failed to read response body: %v", resp.StatusCode, err)
		}
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, respBody.String())
	}

	return nil
}
