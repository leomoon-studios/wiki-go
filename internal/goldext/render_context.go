package goldext

import (
	"sync"

	"github.com/yuin/goldmark/parser"
)

var renderStateKey = parser.NewContextKey()

// RenderState contains extension state belonging to one Markdown conversion.
// It replaces the need for package-level placeholder maps as extensions are
// migrated to trusted nodes.
type RenderState struct {
	DocumentPath string

	mu       sync.RWMutex
	sequence uint64
	values   map[string]any
}

// NewRenderContext creates a fresh Goldmark parser context for one render.
func NewRenderContext(documentPath string) parser.Context {
	context := parser.NewContext()
	context.Set(renderStateKey, &RenderState{
		DocumentPath: documentPath,
		values:       make(map[string]any),
	})
	return context
}

// RenderStateFromContext returns the state associated with a parser context.
// A state is created when an extension is used with a plain Goldmark context.
func RenderStateFromContext(context parser.Context) *RenderState {
	state := context.ComputeIfAbsent(renderStateKey, func() any {
		return &RenderState{values: make(map[string]any)}
	})
	return state.(*RenderState)
}

// NextID returns a render-local normalized identifier.
func (s *RenderState) NextID(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sequence++
	prefix = NormalizeIdentifier(prefix)
	if prefix == "" {
		prefix = "wikigo"
	}
	return prefix + "-" + uintToString(s.sequence)
}

// Set stores render-local extension data.
func (s *RenderState) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
}

// Get retrieves render-local extension data.
func (s *RenderState) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[key]
	return value, ok
}

func uintToString(value uint64) string {
	if value == 0 {
		return "0"
	}

	var digits [20]byte
	position := len(digits)
	for value > 0 {
		position--
		digits[position] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[position:])
}
