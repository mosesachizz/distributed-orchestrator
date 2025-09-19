package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func GenerateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func GetNodeIP() string {
	// Try environment variable first
	if ip := os.Getenv("NODE_IP"); ip != "" {
		return ip
	}

	// Try to get external IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	
	// Handle common suffixes
	if strings.HasSuffix(s, "s") {
		return time.ParseDuration(s)
	}
	if strings.HasSuffix(s, "m") {
		return time.ParseDuration(s)
	}
	if strings.HasSuffix(s, "h") {
		return time.ParseDuration(s)
	}
	
	// Assume seconds if no suffix
	return time.ParseDuration(s + "s")
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, s string) []string {
	result := []string{}
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func StringPtr(s string) *string {
	return &s
}

func BoolPtr(b bool) *bool {
	return &b
}

func IntPtr(i int) *int {
	return &i
}

func NowUnix() int64 {
	return time.Now().Unix()
}

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}