package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"
)

// ANSI color codes.
const (
	colorReset   = "\033[0m"
	colorDim     = "\033[2m"
	colorBold    = "\033[1m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
	colorWhite   = "\033[97m"
	colorOrange  = "\033[38;5;208m" // 256-color orange
)

// ColoredHandler is a slog.Handler that outputs colored text logs.
type ColoredHandler struct {
	opts   slog.HandlerOptions
	writer io.Writer
	mu     *sync.Mutex
	attrs  []slog.Attr
	groups []string
}

// NewColoredHandler creates a new colored console handler.
func NewColoredHandler(w io.Writer, opts *slog.HandlerOptions) *ColoredHandler {
	if opts == nil {
		opts = new(slog.HandlerOptions)
	}
	return &ColoredHandler{
		opts:   *opts,
		writer: w,
		mu:     new(sync.Mutex),
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *ColoredHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Ordered keys for consistent log output.
var orderedKeys = []string{"method", "status", "path", "latency", "request_id", "client_ip"}

// Handle formats and writes a log record with colors.
func (h *ColoredHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Format timestamp (dimmed)
	timeStr := fmt.Sprintf("%s%s%s",
		colorDim+colorGray,
		r.Time.Format(time.DateTime),
		colorReset,
	)

	// Format level with color
	levelStr := h.formatLevel(r.Level)

	// Format message
	msgStr := fmt.Sprintf("%s%s%s", colorWhite, r.Message, colorReset)

	// Build the log line (errors ignored - nothing meaningful to do if stdout fails)
	_, _ = fmt.Fprintf(h.writer, "%s  %s  %s", timeStr, levelStr, msgStr)

	// Collect all attributes
	allAttrs := make(map[string]slog.Attr)
	for _, attr := range h.attrs {
		allAttrs[attr.Key] = attr
	}
	r.Attrs(func(a slog.Attr) bool {
		allAttrs[a.Key] = a
		return true
	})

	// Write ordered keys first
	for _, key := range orderedKeys {
		if attr, ok := allAttrs[key]; ok {
			h.writeAttr(attr)
			delete(allAttrs, key)
		}
	}

	// Write remaining attributes
	for _, attr := range allAttrs {
		h.writeAttr(attr)
	}

	// End line and add spacing
	_, _ = fmt.Fprintln(h.writer)
	_, _ = fmt.Fprintln(h.writer) // Extra blank line for readability

	return nil
}

// WithAttrs returns a new handler with additional attributes.
func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs), len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	newAttrs = append(newAttrs, attrs...)

	return &ColoredHandler{
		opts:   h.opts,
		writer: h.writer,
		mu:     h.mu,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

// WithGroup returns a new handler with a group name.
func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups), len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups = append(newGroups, name)

	return &ColoredHandler{
		opts:   h.opts,
		writer: h.writer,
		mu:     h.mu,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

// formatLevel returns the level string with appropriate color.
func (h *ColoredHandler) formatLevel(level slog.Level) string {
	var color, label string

	switch {
	case level >= slog.LevelError:
		color = colorRed + colorBold
		label = "ERROR"
	case level >= slog.LevelWarn:
		color = colorYellow + colorBold
		label = "WARN "
	case level >= slog.LevelInfo:
		color = colorCyan + colorBold
		label = "INFO "
	default:
		color = colorGray
		label = "DEBUG"
	}

	return fmt.Sprintf("%s%s%s", color, label, colorReset)
}

// writeAttr writes a single attribute with subtle formatting.
func (h *ColoredHandler) writeAttr(a slog.Attr) {
	if a.Equal(slog.Attr{}) {
		return
	}

	key := a.Key
	for _, g := range h.groups {
		key = g + "." + key
	}

	valueColor := h.colorForKey(key, a.Value)

	_, _ = fmt.Fprintf(h.writer, "  %s%s%s=%s%v%s",
		colorGray, key, colorReset,
		valueColor, a.Value.Any(), colorReset,
	)
}

// colorForKey returns the appropriate color for a value based on its key.
func (h *ColoredHandler) colorForKey(key string, value slog.Value) string {
	switch key {
	case "request_id":
		return colorMagenta
	case "method":
		return colorBlue + colorBold
	case "path", "url":
		return colorCyan
	case "status":
		// Color status based on HTTP status code
		if status, ok := value.Any().(int); ok {
			switch {
			case status >= 500:
				return colorRed + colorBold
			case status >= 400:
				return colorYellow
			case status >= 200 && status < 300:
				return colorGreen
			}
		}
		return colorOrange
	case "latency":
		return colorYellow
	case "error", "errors":
		return colorRed
	case "app", "version":
		return colorCyan
	case "client_ip":
		return colorGray
	case "note":
		return colorYellow
	default:
		return colorWhite
	}
}
