package texpr

import (
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/mitchellh/mapstructure"
)

func ConvertMessage(input interface{}) (EnvMessage, error) {
	switch v := input.(type) {
	case *tg.Message:
		return ConvertEnvMessage(v), nil
	case map[string]interface{}:
		return convertJSONMessage(v)
	default:
		// Try to decode as FMessage struct
		var fm struct {
			ID     int         `mapstructure:"id"`
			Type   string      `mapstructure:"type"`
			Date   int         `mapstructure:"date"`
			File   string      `mapstructure:"file"`
			Photo  string      `mapstructure:"photo"`
			FromID string      `mapstructure:"from_id"`
			From   string      `mapstructure:"from"`
			Text   interface{} `mapstructure:"text"`
		}

		if err := mapstructure.WeakDecode(input, &fm); err != nil {
			return EnvMessage{}, fmt.Errorf("unsupported input type: %T", input)
		}

		return convertFMessage(fm)
	}
}

func convertJSONMessage(data map[string]interface{}) (EnvMessage, error) {
	var msg struct {
		ID     int         `mapstructure:"id"`
		Type   string      `mapstructure:"type"`
		Date   int         `mapstructure:"date"`
		File   string      `mapstructure:"file"`
		Photo  string      `mapstructure:"photo"`
		FromID string      `mapstructure:"from_id"`
		From   string      `mapstructure:"from"`
		Text   interface{} `mapstructure:"text"`
	}

	if err := mapstructure.WeakDecode(data, &msg); err != nil {
		return EnvMessage{}, err
	}

	return convertFMessage(msg)
}

func convertFMessage(msg struct {
	ID     int         `mapstructure:"id"`
	Type   string      `mapstructure:"type"`
	Date   int         `mapstructure:"date"`
	File   string      `mapstructure:"file"`
	Photo  string      `mapstructure:"photo"`
	FromID string      `mapstructure:"from_id"`
	From   string      `mapstructure:"from"`
	Text   interface{} `mapstructure:"text"`
}) (EnvMessage, error) {
	env := EnvMessage{
		ID:   msg.ID,
		Date: msg.Date,
	}

	env.Message = extractJSONText(msg.Text)

	// Set media info
	if msg.File != "" {
		env.Media.Name = msg.File
	} else if msg.Photo != "" {
		env.Media.Name = msg.Photo
	}

	if msg.FromID != "" {
		var id int64
		if strings.HasPrefix(msg.FromID, "user") {
			fmt.Sscanf(msg.FromID, "user%d", &id)
		} else if strings.HasPrefix(msg.FromID, "channel") {
			fmt.Sscanf(msg.FromID, "channel%d", &id)
		}
		env.FromID = id
	}

	return env, nil
}

func extractJSONText(raw interface{}) string {
	switch v := raw.(type) {
	case string:
		return v
	case []interface{}:
		var buf strings.Builder
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if text, ok := m["text"].(string); ok {
					buf.WriteString(text)
				}
			} else if t, ok := item.(string); ok {
				buf.WriteString(t)
			}
		}
		return buf.String()
	}
	return ""
}
