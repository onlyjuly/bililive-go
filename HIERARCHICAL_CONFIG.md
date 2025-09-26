# Hierarchical Configuration System Documentation

## Overview

The bililive-go application now supports a three-level hierarchical configuration system that allows for granular control over recording settings:

1. **Global Level**: Default settings that apply to all live rooms
2. **Platform Level**: Settings that apply to all rooms on a specific platform (e.g., Bilibili, Douyin)
3. **Room Level**: Settings that apply to individual live rooms

Settings at lower levels override higher-level settings, following the priority: **Room Level > Platform Level > Global Level**

## Configuration Structure

### Overridable Settings

The following settings can be overridden at platform and room levels:

- `interval`: Check interval in seconds
- `out_put_path`: Output directory for recordings
- `ffmpeg_path`: Path to FFmpeg executable
- `log`: Logging configuration
- `feature`: Feature flags (e.g., native FLV parser)
- `out_put_tmpl`: Output filename template
- `video_split_strategies`: Video splitting configuration
- `on_record_finished`: Post-recording actions
- `timeout_in_us`: Timeout configuration

### Platform-Specific Settings

Each platform can have additional settings:

- `name`: Human-readable platform name
- `min_access_interval_sec`: Minimum time between API requests to prevent rate limiting

## Example Configuration

```yaml
# Global settings
interval: 30
out_put_path: ./recordings
ffmpeg_path: ""

# Platform-specific configurations
platform_configs:
  # Douyin platform settings
  douyin:
    name: "抖音"
    min_access_interval_sec: 5   # Rate limiting: max 1 request per 5 seconds
    interval: 15                 # Override global: check every 15 seconds
    out_put_path: ./recordings/douyin
  
  # Bilibili platform settings  
  bilibili:
    name: "哔哩哔哩"
    min_access_interval_sec: 3
    interval: 20
    feature:
      use_native_flv_parser: true

# Live rooms with room-level overrides
live_rooms:
  - url: https://live.douyin.com/123456789
    is_listening: true
    interval: 10                 # Override platform setting
    
  - url: https://live.bilibili.com/987654321
    is_listening: true
    out_put_path: ./recordings/special  # Override both global and platform
    ffmpeg_path: /opt/ffmpeg/bin/ffmpeg
```

## Web Interface

The web configuration interface supports both GUI and text editing modes:

### GUI Mode
- Visual forms for common settings
- Platform selection dropdown
- Tooltips explaining setting hierarchy
- Input validation

### Text Mode
- Direct YAML editing with syntax highlighting
- Fallback for advanced configuration
- Full access to all configuration options

## API Changes

### New Functions

```go
// Get resolved configuration for a specific room
func (c *Config) ResolveConfigForRoom(room *LiveRoom, platformName string) ResolvedConfig

// Get platform-specific access rate limit
func (c *Config) GetPlatformMinAccessInterval(platformName string) int

// Get effective configuration for a room URL
func (c *Config) GetEffectiveConfigForRoom(roomUrl string) ResolvedConfig

// Map URL to platform key
func GetPlatformKeyFromUrl(urlStr string) string
```

### Backward Compatibility

The system maintains full backward compatibility with existing configuration files. Old configurations will work unchanged, with new hierarchical features available as opt-in additions.

## Platform Rate Limiting

The system now supports platform-level rate limiting to prevent being blocked by streaming services:

- Each platform can have a `min_access_interval_sec` setting
- The system enforces minimum time between API requests per platform
- Helps prevent triggering anti-bot measures

## Supported Platforms

The following platforms are supported with automatic URL-to-platform mapping:

- Bilibili (`live.bilibili.com`)
- Douyin (`live.douyin.com`, `v.douyin.com`)
- Douyu (`www.douyu.com`)
- Kuaishou (`live.kuaishou.com`)
- YY (`www.yy.com`)
- AcFun (`live.acfun.cn`)
- And many more...

## Migration Guide

### From Old Configuration

Existing configurations require no changes. The system will:

1. Load existing configurations normally
2. Initialize empty platform_configs if not present
3. Apply global settings to all rooms as before

### Adding Hierarchical Configuration

To use the new features:

1. Add `platform_configs` section to your YAML
2. Define platform-specific settings under each platform key
3. Add room-level overrides directly to live room entries

### Example Migration

**Before:**
```yaml
interval: 30
out_put_path: ./
live_rooms:
  - url: https://live.bilibili.com/123456
```

**After (with hierarchical config):**
```yaml
interval: 30
out_put_path: ./

platform_configs:
  bilibili:
    name: "哔哩哔哩"
    min_access_interval_sec: 3
    interval: 20

live_rooms:
  - url: https://live.bilibili.com/123456
    interval: 15  # Room-specific override
```

## Testing

The implementation includes comprehensive tests:

- Configuration parsing and resolution
- Backward compatibility
- Platform mapping
- GUI/text mode integration
- Hierarchical override logic

Run tests with:
```bash
go test ./src/configs/
```