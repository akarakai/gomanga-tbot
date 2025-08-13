package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"os"
)

func setSLogger() {
	handler := NewClassicHandler(os.Stdout, slog.LevelDebug)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

type classicHandler struct {
	w     io.Writer
	level slog.Level
}

func (h *classicHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *classicHandler) Handle(_ context.Context, r slog.Record) error {
	// Format time
	t := r.Time.Format("2006/01/02 15:04:05")
	
	// Format level
	level := fmt.Sprintf("[%s]", r.Level.String())
	
	// Get the correct caller information (skip the slog internal calls)
	var source string
	fs := runtime.CallersFrames([]uintptr{r.PC})
	frame, _ := fs.Next()
	if frame.PC != 0 {
		file := filepath.Base(frame.File)
		funcName := filepath.Base(strings.TrimPrefix(frame.Function, "main."))
		source = fmt.Sprintf("%s %s %d", file, funcName, frame.Line)
	}
	
	// Combine all parts
	logLine := fmt.Sprintf("%s %s %s %s\n", t, level, source, r.Message)
	
	_, err := io.WriteString(h.w, logLine)
	return err
}

func (h *classicHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *classicHandler) WithGroup(name string) slog.Handler {
	return h
}

func NewClassicHandler(w io.Writer, level slog.Level) *classicHandler {
	return &classicHandler{
		w:     w,
		level: level,
	}
}
