package configs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	file := "../../config.yml"
	c, err := NewConfigWithFile("../../config.yml")
	assert.NoError(t, err)
	assert.Equal(t, file, c.File)
}

func TestRPC_Verify(t *testing.T) {
	var rpc *RPC
	assert.NoError(t, rpc.verify())
	rpc = new(RPC)
	rpc.Bind = "foo@bar"
	assert.NoError(t, rpc.verify())
	rpc.Enable = true
	assert.Error(t, rpc.verify())
}

func TestConfig_Verify(t *testing.T) {
	var cfg *Config
	assert.Error(t, cfg.Verify())
	cfg = &Config{
		RPC:        defaultRPC,
		Interval:   30,
		OutPutPath: os.TempDir(),
	}
	assert.NoError(t, cfg.Verify())
	cfg.Interval = 0
	assert.Error(t, cfg.Verify())
	cfg.Interval = 30
	cfg.OutPutPath = "foobar"
	assert.Error(t, cfg.Verify())
	cfg.OutPutPath = os.TempDir()
	cfg.RPC.Enable = false
	assert.Error(t, cfg.Verify())
}

func TestResolveConfigForRoom(t *testing.T) {
	cfg := &Config{
		Interval:   60,
		OutPutPath: "/global",
		FfmpegPath: "/usr/bin/ffmpeg",
		PlatformConfigs: map[string]PlatformConfig{
			"douyin": {
				OverridableConfig: OverridableConfig{
					Interval:   intPtr(30),
					OutPutPath: stringPtr("/douyin"),
				},
			},
		},
	}

	room := &LiveRoom{
		Url: "https://live.douyin.com/123456",
		OverridableConfig: OverridableConfig{
			Interval: intPtr(15),
		},
	}

	resolved := cfg.ResolveConfigForRoom(room, "douyin")
	
	// Room-level override should take precedence
	assert.Equal(t, 15, resolved.Interval)
	// Platform-level override should take precedence over global
	assert.Equal(t, "/douyin", resolved.OutPutPath)
	// Global value should be used when no override exists
	assert.Equal(t, "/usr/bin/ffmpeg", resolved.FfmpegPath)
}

func TestGetPlatformMinAccessInterval(t *testing.T) {
	cfg := &Config{
		PlatformConfigs: map[string]PlatformConfig{
			"douyin": {
				OverridableConfig:    OverridableConfig{},
				MinAccessIntervalSec: 5,
			},
		},
	}

	// Test existing platform
	interval := cfg.GetPlatformMinAccessInterval("douyin")
	assert.Equal(t, 5, interval)

	// Test non-existing platform
	interval = cfg.GetPlatformMinAccessInterval("bilibili")
	assert.Equal(t, 0, interval)
}

func TestBackwardsCompatibility(t *testing.T) {
	// Test that old config files still work
	oldConfigYaml := `
rpc:
  enable: true
  bind: :8080
debug: false
interval: 30
out_put_path: ./
live_rooms:
- url: https://live.bilibili.com/123456
  is_listening: true
`
	cfg, err := NewConfigWithBytes([]byte(oldConfigYaml))
	assert.NoError(t, err)
	assert.NotNil(t, cfg.PlatformConfigs)
	assert.Equal(t, 30, cfg.Interval)
	assert.Len(t, cfg.LiveRooms, 1)
	assert.Equal(t, "https://live.bilibili.com/123456", cfg.LiveRooms[0].Url)
	
	// Test that resolve works with no overrides
	resolved := cfg.ResolveConfigForRoom(&cfg.LiveRooms[0], "bilibili")
	assert.Equal(t, 30, resolved.Interval)
	assert.Equal(t, "./", resolved.OutPutPath)
}

// Helper functions for pointer conversion
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
