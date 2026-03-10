package runtime

import (
	"bytes"
	"log"
	"strings"
	"sync"
)

type processLogWriter struct {
	mu      sync.Mutex
	service string
	part    string
	stream  string
	buffer  bytes.Buffer
}

func newProcessLogWriter(service, part, stream string) *processLogWriter {
	return &processLogWriter{service: service, part: part, stream: stream}
}

func (w *processLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.buffer.Write(p); err != nil {
		return 0, err
	}
	for {
		line, err := w.buffer.ReadString('\n')
		if err != nil {
			w.buffer.WriteString(line)
			return len(p), nil
		}
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			log.Printf("[%s/%s/%s] %s", w.service, w.part, w.stream, trimmed)
		}
	}
}
