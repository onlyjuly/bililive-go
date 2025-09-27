package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bililive-go/bililive-go/src/pkg/media"
	"github.com/bililive-go/bililive-go/src/live"
	"github.com/bililive-go/bililive-go/src/configs"
	"github.com/bililive-go/bililive-go/src/types"
	"github.com/sirupsen/logrus"
)

// Simple live implementation for demo
type DemoLive struct {
	platformName string
}

func (d *DemoLive) GetPlatformCNName() string { return d.platformName }
func (d *DemoLive) GetInfo() (*live.Info, error) { return nil, nil }
func (d *DemoLive) GetStreamInfos() ([]*live.StreamUrlInfo, error) { return nil, nil }
func (d *DemoLive) GetStreamUrls() ([]*url.URL, error) { return nil, nil }
func (d *DemoLive) GetRawUrl() string { return "" }
func (d *DemoLive) GetLiveId() types.LiveID { return "" }
func (d *DemoLive) SetLiveIdByString(string) {}
func (d *DemoLive) GetLastStartTime() time.Time { return time.Time{} }
func (d *DemoLive) SetLastStartTime(time.Time) {}
func (d *DemoLive) GetOptions() *live.Options { return nil }
func (d *DemoLive) UpdateLiveOptionsbyConfig(context.Context, *configs.LiveRoom) error { return nil }

func main() {
	// Create demo live info
	demoLive := &DemoLive{platformName: "Bilibili"}
	
	info := &live.Info{
		Live:     demoLive,
		HostName: "测试主播",
		RoomName: "精彩直播间",
		Status:   true,
	}

	logger := logrus.NewEntry(logrus.New())
	logger.Logger.SetLevel(logrus.InfoLevel)

	// Use the test video we created earlier
	videoPath := "/tmp/test_video.mp4"
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		log.Fatal("Test video not found at /tmp/test_video.mp4")
	}

	fmt.Println("=== Demo: NFO and Poster Generation ===")
	fmt.Printf("Video file: %s\n", videoPath)
	fmt.Printf("Host: %s\n", info.HostName)
	fmt.Printf("Room: %s\n", info.RoomName)
	fmt.Printf("Platform: %s\n", info.Live.GetPlatformCNName())
	fmt.Println()

	// Generate NFO
	fmt.Println("Generating NFO file...")
	if err := media.GenerateNFO(videoPath, info, logger); err != nil {
		log.Fatalf("Failed to generate NFO: %v", err)
	}

	// Generate poster
	fmt.Println("Generating poster image...")
	if err := media.GeneratePoster(videoPath, "ffmpeg", "00:00:02", logger); err != nil {
		log.Fatalf("Failed to generate poster: %v", err)
	}

	fmt.Println("\n=== Generated Files ===")
	
	// Check and display NFO
	nfoPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + ".nfo"
	if content, err := os.ReadFile(nfoPath); err == nil {
		fmt.Printf("NFO file created: %s (%d bytes)\n", nfoPath, len(content))
		fmt.Println("NFO content preview:")
		fmt.Println("---")
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if i < 10 { // Show first 10 lines
				fmt.Println(line)
			}
		}
		if len(lines) > 10 {
			fmt.Println("... (content truncated)")
		}
		fmt.Println("---")
	} else {
		fmt.Printf("Failed to read NFO file: %v\n", err)
	}

	// Check poster
	posterPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + ".jpg"
	if info, err := os.Stat(posterPath); err == nil {
		fmt.Printf("Poster image created: %s (%d bytes)\n", posterPath, info.Size())
	} else {
		fmt.Printf("Failed to stat poster file: %v\n", err)
	}

	fmt.Println("\n=== Demo Completed Successfully! ===")
}