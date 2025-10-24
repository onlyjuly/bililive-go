package ntfy

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// SendMessage 发送ntfy消息
func SendMessage(url, token, tag, hostname, platform, liveURL, schemeURL string) error {
	// 构造消息内容
	message := fmt.Sprintf("%s开播", platform)
	title := hostname

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, strings.NewReader(message))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置必要的请求头
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Title", title)
	req.Header.Set("Tags", tag)

	// 如果提供了scheme URL，则设置Click头
	if schemeURL != "" {
		req.Header.Set("Click", schemeURL)
	}

	// 设置Actions头，用于打开直播间
	if liveURL != "" {
		// 确保liveURL有https://前缀
		fullURL := liveURL
		if !strings.HasPrefix(liveURL, "http://") && !strings.HasPrefix(liveURL, "https://") {
			fullURL = "https://" + liveURL
		}
		action := fmt.Sprintf(`[{"action":"view","label":"打开直播间","url":"%s"}]`, fullURL)
		req.Header.Set("Actions", action)
	}

	// 如果提供了token，则添加Authorization头
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetSchemeURL 从scheme_config.ini文件中根据hostname获取对应的scheme URL
func GetSchemeURL(hostname string) string {
	// 读取scheme_config.ini文件
	data, err := os.ReadFile("scheme_config.ini")
	if err != nil {
		// 如果文件不存在或无法读取，返回空字符串
		return ""
	}

	// 按行分割文件内容
	lines := strings.Split(string(data), "\n")
	
	// 遍历每一行查找匹配的hostname
	for _, line := range lines {
		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// 分割hostname和scheme URL
		parts := strings.Split(line, ",")
		if len(parts) >= 2 && strings.TrimSpace(parts[0]) == hostname {
			return strings.TrimSpace(parts[1])
		}
	}
	
	return ""
}