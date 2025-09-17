package osdetect

import (
	"runtime"
	"strings"
)

type OSType string

const (
	Windows OSType = "windows"
	Linux   OSType = "linux"
	Darwin  OSType = "darwin"
	Unknown OSType = "unknown"
)

func DetectOS() OSType {
	switch strings.ToLower(runtime.GOOS) {
	case "windows":
		return Windows
	case "linux":
		return Linux
	case "darwin":
		return Darwin
	default:
		return Unknown
	}
}

func IsWindows() bool {
	return DetectOS() == Windows
}

func IsLinux() bool {
	return DetectOS() == Linux
}

func IsDarwin() bool {
	return DetectOS() == Darwin
}

func GetOSString() string {
	return string(DetectOS())
}