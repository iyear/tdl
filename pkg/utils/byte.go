package utils

import (
	"fmt"
	"strconv"
	"strings"
)

type _byte struct{}

var Byte _byte

func (b _byte) FormatBinaryBytes(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	if n < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(n)/1024)
	}
	if n < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(n)/1024/1024)
	}
	if n < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2f GB", float64(n)/1024/1024/1024)
	}
	return fmt.Sprintf("%.2f TB", float64(n)/1024/1024/1024/1024)
}

func (b _byte) ParseBinaryBytes(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}

	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	var multiplier int64 = 1
	if strings.HasSuffix(s, "TB") {
		multiplier = 1024 * 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "TB")
	} else if strings.HasSuffix(s, "GB") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "KB") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "B") {
		s = strings.TrimSuffix(s, "B")
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return int64(val * float64(multiplier)), nil
}
