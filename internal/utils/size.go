package utils

import "fmt"

func FormatSize(size int64) string {
	if size <= 0 {
		return "0 B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(size)
	index := 0

	for value >= 1024 && index < len(units)-1 {
		value /= 1024
		index++
	}

	return fmt.Sprintf("%.2f %s", value, units[index])
}
