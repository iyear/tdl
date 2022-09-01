package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type telegram struct{}

var Telegram telegram

// ParseMessageLink return dialog id, msg id, error
func (t telegram) ParseMessageLink(s string) (int64, int, error) {
	u, err := url.Parse(s)
	if err != nil {
		return 0, 0, err
	}

	paths := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")

	if len(paths) != 3 {
		return 0, 0, fmt.Errorf("invalid link path: %s", paths)
	}

	dialog, err := strconv.ParseInt(paths[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	msg, err := strconv.Atoi(paths[2])
	if err != nil {
		return 0, 0, err
	}

	return dialog, msg, nil
}
