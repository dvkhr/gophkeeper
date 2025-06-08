// pkg/logger/handler/text.go

package handler

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"log/slog"

	"golang.org/x/term"
)

type TextHandler struct {
	out       io.Writer
	level     slog.Level
	isColored bool
}

func NewTextHandler(out io.Writer, level slog.Level) *TextHandler {
	isColored := false
	if f, ok := out.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		isColored = true
	}
	return &TextHandler{
		out:       out,
		level:     level,
		isColored: isColored,
	}
}

func (h *TextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *TextHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()

	if h.isColored {
		switch level {
		case "DEBUG":
			fmt.Fprintf(h.out, "\033[34m") // Синий
		case "WARN":
			fmt.Fprintf(h.out, "\033[33m") // Жёлтый
		case "ERROR":
			fmt.Fprintf(h.out, "\033[31m") // Красный
		}
	}

	fmt.Fprintf(h.out, "%s [%s] %s",
		r.Time.Format(time.RFC3339),
		level,
		r.Message,
	)

	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(h.out, " %s=%v", a.Key, a.Value)
		return true
	})

	fmt.Fprintln(h.out)
	if h.isColored {
		fmt.Fprintf(h.out, "\033[0m")
	}
	return nil
}

func (h *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *TextHandler) WithGroup(name string) slog.Handler {
	return h
}
