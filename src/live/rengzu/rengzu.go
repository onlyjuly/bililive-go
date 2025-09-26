// Package rengzu provides live streaming support for 时光直播 (Rengzu.com).
// This package implements the Live interface to handle room information
// extraction and stream URL discovery for the Rengzu.com live streaming platform.
//
// Supported URL format: https://www.rengzu.com/{room_id}
// Example: https://www.rengzu.com/191996
package rengzu

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/bililive-go/bililive-go/src/live"
	"github.com/bililive-go/bililive-go/src/live/internal"
	"github.com/bililive-go/bililive-go/src/pkg/utils"
)

const (
	domain = "www.rengzu.com"
	cnName = "时光直播"

	// Common patterns for live streaming platforms
	apiBaseUrl = "https://www.rengzu.com/api"
)

func init() {
	live.Register(domain, new(builder))
}

type builder struct{}

func (b *builder) Build(url *url.URL) (live.Live, error) {
	return &Live{
		BaseLive: internal.NewBaseLive(url),
	}, nil
}

type Live struct {
	internal.BaseLive
}

// getRoomID extracts room ID from URL path
// URL format: https://www.rengzu.com/191996
func (l *Live) getRoomID() string {
	path := strings.Trim(l.Url.Path, "/")
	if path == "" {
		return ""
	}
	// Return the last part of the path as room ID
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	roomID := l.getRoomID()
	if roomID == "" {
		return nil, live.ErrRoomNotExist
	}

	// Try to get page content first to extract room information
	resp, err := l.RequestSession.Get(l.Url.String(), live.CommonUserAgent)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, live.ErrRoomNotExist
	}

	body, err := resp.Text()
	if err != nil {
		return nil, err
	}

	// Look for common patterns in the HTML that might contain room info
	// These are common patterns found in live streaming sites
	var hostName, roomName string
	var isLive bool

	// Try to extract title from HTML title tag
	titleRe := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	if matches := titleRe.FindStringSubmatch(body); len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		if title != "" {
			roomName = title
		}
	}

	// Look for JSON data that might contain room information
	// Common patterns: window.__INITIAL_DATA__, window.__NUXT__, etc.
	jsonPatterns := []string{
		`window\.__INITIAL_DATA__\s*=\s*({.*?});`,
		`window\.__NUXT__\s*=\s*({.*?});`,
		`window\.initData\s*=\s*({.*?});`,
		`"roomInfo"\s*:\s*({.*?})`,
	}

	for _, pattern := range jsonPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(body); len(matches) > 1 {
			jsonStr := matches[1]
			result := gjson.Parse(jsonStr)

			// Try to extract common fields
			if result.IsObject() {
				// Common field names in live streaming platforms
				possibleHostNames := []string{"anchor_name", "host_name", "nickname", "username", "anchor.name", "user.nickname"}
				possibleRoomNames := []string{"title", "room_title", "live_title", "room_name", "room.title"}
				possibleStatus := []string{"status", "is_live", "live_status", "online"}

				for _, field := range possibleHostNames {
					if val := result.Get(field).String(); val != "" {
						hostName = val
						break
					}
				}

				for _, field := range possibleRoomNames {
					if val := result.Get(field).String(); val != "" && roomName == "" {
						roomName = val
						break
					}
				}

				for _, field := range possibleStatus {
					if result.Get(field).Exists() {
						status := result.Get(field)
						if status.Type == gjson.Number {
							isLive = status.Int() == 1 || status.Int() == 2
						} else if status.Type == gjson.True {
							isLive = true
						} else if status.Type == gjson.String {
							statusStr := strings.ToLower(status.String())
							isLive = statusStr == "live" || statusStr == "online" || statusStr == "1"
						}
						break
					}
				}

				if hostName != "" || roomName != "" {
					break
				}
			}
		}
	}

	// If we couldn't find proper names, use default values
	if hostName == "" {
		hostName = fmt.Sprintf("主播_%s", roomID)
	}
	if roomName == "" {
		roomName = fmt.Sprintf("直播间_%s", roomID)
	}

	// Check if the page indicates the stream is offline
	offlineIndicators := []string{
		"直播已结束", "主播不在线", "未开播", "offline", "not live", "ended",
	}
	for _, indicator := range offlineIndicators {
		if strings.Contains(strings.ToLower(body), strings.ToLower(indicator)) {
			isLive = false
			break
		}
	}

	info = &live.Info{
		Live:     l,
		HostName: hostName,
		RoomName: roomName,
		Status:   isLive,
	}
	return info, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	roomID := l.getRoomID()
	if roomID == "" {
		return nil, live.ErrRoomNotExist
	}

	// First try to get the main page content
	resp, err := l.RequestSession.Get(l.Url.String(), live.CommonUserAgent)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, live.ErrRoomNotExist
	}

	body, err := resp.Text()
	if err != nil {
		return nil, err
	}

	// Look for stream URLs in common patterns
	streamUrlPatterns := []string{
		// Common patterns for stream URLs
		`"stream_url"\s*:\s*"([^"]+)"`,
		`"play_url"\s*:\s*"([^"]+)"`,
		`"hls_url"\s*:\s*"([^"]+)"`,
		`"flv_url"\s*:\s*"([^"]+)"`,
		`"rtmp_url"\s*:\s*"([^"]+)"`,
		// Direct URL patterns
		`https?://[^"'\s]+\.m3u8[^"'\s]*`,
		`https?://[^"'\s]+\.flv[^"'\s]*`,
		`rtmp://[^"'\s]+[^"'\s]*`,
	}

	var streamUrls []string
	for _, pattern := range streamUrlPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(body, -1)
		for _, match := range matches {
			if len(match) > 1 {
				streamUrls = append(streamUrls, match[1])
			} else if len(match) > 0 {
				streamUrls = append(streamUrls, match[0])
			}
		}
	}

	// Try API endpoints that are common in live streaming platforms
	apiEndpoints := []string{
		fmt.Sprintf("/api/room/%s/stream", roomID),
		fmt.Sprintf("/api/live/%s", roomID),
		fmt.Sprintf("/room/stream?room_id=%s", roomID),
		fmt.Sprintf("/live/play?id=%s", roomID),
	}

	for _, endpoint := range apiEndpoints {
		apiUrl := fmt.Sprintf("https://%s%s", l.Url.Host, endpoint)
		resp, err := l.RequestSession.Get(apiUrl, live.CommonUserAgent)
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			continue
		}

		apiBody, err := resp.Text()
		if err != nil {
			continue
		}

		// Try to parse as JSON
		result := gjson.Parse(apiBody)
		if result.IsObject() {
			// Look for stream URLs in API response
			streamFields := []string{"stream_url", "play_url", "hls_url", "flv_url", "rtmp_url", "data.stream_url", "data.play_url"}
			for _, field := range streamFields {
				if url := result.Get(field).String(); url != "" {
					streamUrls = append(streamUrls, url)
				}
			}
		} else {
			// Try regex patterns on API response
			for _, pattern := range streamUrlPatterns {
				re := regexp.MustCompile(pattern)
				matches := re.FindAllStringSubmatch(apiBody, -1)
				for _, match := range matches {
					if len(match) > 1 {
						streamUrls = append(streamUrls, match[1])
					}
				}
			}
		}
	}

	// Remove duplicates and invalid URLs
	seen := make(map[string]bool)
	var validUrls []string
	for _, streamUrl := range streamUrls {
		if streamUrl == "" || seen[streamUrl] {
			continue
		}
		seen[streamUrl] = true

		// Basic URL validation
		if strings.HasPrefix(streamUrl, "http") || strings.HasPrefix(streamUrl, "rtmp") {
			validUrls = append(validUrls, streamUrl)
		}
	}

	if len(validUrls) == 0 {
		return nil, fmt.Errorf("no stream URLs found for room %s", roomID)
	}

	return utils.GenUrls(validUrls...)
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
