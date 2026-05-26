package services

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const MaxRecurDepth = 3

type Environ struct {
	storage map[string]Entry
}

func NewWithFile(filename string) (*Environ, error) {
	env := &Environ{storage: make(map[string]Entry)}
	err := env.Load(os.Open(filename))
	return env, err
}

func (v *Environ) Load(reader io.ReadCloser, err error) error {
	if err != nil || reader == nil {
		return err
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			err = closeErr
		}
	}()
	return v.ScanLines(reader)
}

func (v *Environ) ScanLines(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimLeftFunc(scanner.Text(), unicode.IsSpace)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimRightFunc(parts[0], unicode.IsSpace)
		value := strings.TrimSpace(parts[1])
		if len(value) >= 2 {
			firstChar, lastChar := value[0], value[len(value)-1]
			if (firstChar == '"' || firstChar == '\'') && (firstChar == lastChar) {
				value = value[1 : len(value)-1]
			}
		}
		v.storage[key] = Entry{Key: key, Value: value}
	}
	return scanner.Err()
}

func (v *Environ) Lookup(key string) (Entry, bool) {
	if entry, ok := v.storage[key]; ok {
		return entry, true
	}
	ee := os.Getenv(key)
	if ee != "" {
		entry := Entry{Key: key, Value: ee}
		v.storage[key] = entry
		return entry, true
	}
	return Entry{}, false
}

func (v *Environ) recurGet(key string, depth int) string {
	s := v.GetStr(key)
	if s != "" && depth <= MaxRecurDepth && strings.Contains(s, "$") {
		return os.Expand(s, func(k string) string {
			return v.recurGet(k, depth+1)
		})
	}
	return s
}

func (v *Environ) Get(key string) string {
	return v.recurGet(key, 1)
}

func (v *Environ) GetStr(key string, fallback ...string) string {
	if entry, ok := v.Lookup(key); ok || len(fallback) == 0 {
		return entry.Str()
	}
	return fallback[0]
}

func (v *Environ) GetInt(key string, fallback ...int) int {
	if entry, ok := v.Lookup(key); ok || len(fallback) == 0 {
		return entry.Int()
	}
	return fallback[0]
}

type Entry struct {
	Key   string
	Value string
}

func (t *Entry) Str() string {
	return t.Value
}

func (t *Entry) Int() int {
	if val, err := strconv.Atoi(t.Value); err == nil {
		return val
	}
	return 0
}
