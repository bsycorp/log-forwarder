package main

import (
	"time"
)

const (
	// Limits to trigger a buffer flush
	MaxBufferBytes = 100 * 1024
	MaxBufferAge   = time.Duration(15 * time.Second)
)

type LogBuffer struct {
	Messages   []string
	TotalBytes int
	Age        time.Time
	Metadata   MetadataValues
}

func (buf *LogBuffer) Append(msg string) {
	if len(buf.Messages) == 0 {
		buf.Age = time.Now()
	}
	buf.Messages = append(buf.Messages, msg)
	buf.TotalBytes += len(msg)
}

func (buf *LogBuffer) Clear() {
	buf.Messages = []string{}
	buf.TotalBytes = 0
	buf.Age = time.Time{}
}

func (buf *LogBuffer) NeedsFlush() bool {
	bufferAge := time.Since(buf.Age)
	r := len(buf.Messages) > 0 && (buf.TotalBytes > MaxBufferBytes || bufferAge > MaxBufferAge)
	return r
}

func (buf *LogBuffer) GetMessages() []string {
	return buf.Messages
}
