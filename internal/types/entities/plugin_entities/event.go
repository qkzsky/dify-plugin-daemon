package plugin_entities

import "encoding/json"

type PluginUniversalEvent struct {
	Event     string          `json:"event"`
	SessionId string          `json:"session_id"`
	Data      json.RawMessage `json:"data"`
}

const (
	PLUGIN_EVENT_LOG     = "log"
	PLUGIN_EVENT_SESSION = "session"
	PLUGIN_EVENT_ERROR   = "error"
)

type PluginLogEvent struct {
	Level     string  `json:"level"`
	Message   string  `json:"message"`
	Timestamp float64 `json:"timestamp"`
}

type SessionMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

const (
	SESSION_MESSAGE_TYPE_STREAM = "stream"
	SESSION_MESSAGE_TYPE_END    = "end"
	SESSION_MESSAGE_TYPE_INVOKE = "invoke"
)

type InvokeToolResponseChunk struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type InvokeModelResponseChunk struct {
}
