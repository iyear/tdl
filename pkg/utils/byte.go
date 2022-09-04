package utils

import "fmt"

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
