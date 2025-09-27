#!/bin/bash

# Demo script showing NFO and poster generation functionality

echo "=== BiliLive-go NFO and Poster Generation Demo ==="
echo

# Create a test video
echo "Creating test video..."
ffmpeg -f lavfi -i testsrc=duration=5:size=640x480:rate=25 -f lavfi -i sine=frequency=1000:duration=5 \
       -c:v libx264 -t 5 -pix_fmt yuv420p -y demo_live.mp4 >/dev/null 2>&1

# Create test config with NFO and poster generation enabled
cat > demo_config.yml << 'EOF'
rpc:
  enable: false
debug: false
interval: 20
out_put_path: ./demo_output/
ffmpeg_path: 
log:
  out_put_folder: ./
  save_last_log: true
  save_every_log: false
feature:
  use_native_flv_parser: false
live_rooms: []
out_put_tmpl: ''
video_split_strategies:
  on_room_name_changed: false
  max_duration: 0s
  max_file_size: 0
cookies: {}
on_record_finished:
  convert_to_mp4: false
  delete_flv_after_convert: false
  # Enable NFO and poster generation
  generate_nfo: true
  generate_poster: true  
  poster_time: "00:00:02"
  custom_commandline: ""
timeout_in_us: 60000000
EOF

echo "Test video created: demo_live.mp4"
echo

# Create test directory structure
mkdir -p demo_output

# Copy the test video to the output directory to simulate a recorded stream
cp demo_live.mp4 demo_output/

# Now test the media generation functions manually with Go
cat > test_media_generation.go << 'EOF'
package main

import (
    "log"
    "github.com/bililive-go/bililive-go/src/pkg/media"
    "github.com/bililive-go/bililive-go/src/live/mock"
    "github.com/bililive-go/bililive-go/src/live"
    "github.com/sirupsen/logrus"
    "go.uber.org/mock/gomock"
)

func main() {
    ctrl := gomock.NewController(&testingT{})
    defer ctrl.Finish()
    
    mockLive := mock.NewMockLive(ctrl)
    mockLive.EXPECT().GetPlatformCNName().Return("Bilibili").AnyTimes()
    
    info := &live.Info{
        Live:     mockLive,
        HostName: "测试主播",
        RoomName: "精彩直播间",
        Status:   true,
    }
    
    logger := logrus.NewEntry(logrus.New())
    
    filePath := "./demo_output/demo_live.mp4"
    
    // Generate NFO
    err := media.GenerateNFO(filePath, info, logger)
    if err != nil {
        log.Fatalf("Failed to generate NFO: %v", err)
    }
    
    // Generate poster
    err = media.GeneratePoster(filePath, "ffmpeg", "00:00:02", logger)
    if err != nil {
        log.Fatalf("Failed to generate poster: %v", err)
    }
    
    log.Println("Demo completed successfully!")
}

type testingT struct{}
func (t *testingT) Errorf(format string, args ...interface{}) { log.Printf(format, args...) }
func (t *testingT) FailNow() { log.Fatal("Test failed") }
func (t *testingT) Helper() {}
EOF

echo "Running NFO and poster generation demo..."
go mod init demo 2>/dev/null || true
go mod edit -replace github.com/bililive-go/bililive-go=./
go run test_media_generation.go

echo
echo "Generated files:"
ls -la demo_output/demo_live.*

echo
echo "NFO file content:"
echo "=================="
cat demo_output/demo_live.nfo
echo
echo "=================="

echo "Poster image info:"
file demo_output/demo_live.jpg

echo
echo "Demo completed! Check the demo_output directory for generated files."

# Cleanup
rm -f test_media_generation.go demo_config.yml demo_live.mp4 go.mod go.sum
rm -rf demo_output/