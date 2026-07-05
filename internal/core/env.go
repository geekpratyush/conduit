package core

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"sync"
)

// EnvironmentService resolves ${VAR} references in user-entered strings at
// send/connect time:
//
//   - named environments (dev/staging/prod) with one active selection,
//   - a .env file layered under the active environment,
//   - system-environment fallback,
//   - ${VAR:-default}, nested resolution, and \${...} escaping,
//   - secret masking for display/log scrubbing.
//
// Resolution order for ${VAR}: active environment → .env → os.Getenv.
type EnvironmentService struct {
	mu       sync.RWMutex
	envs     map[string]map[string]string // environment name -> vars
	active   string
	dotenv   map[string]string
	secrets  map[string]bool // var names whose values must be masked
}

// varPattern matches ${NAME} and ${NAME:-default}, with an optional preceding
// backslash captured so an escaped \${...} can be emitted literally.
var varPattern = regexp.MustCompile(`(\\?)\$\{([A-Za-z_][A-Za-z0-9_]*)(?::-([^}]*))?\}`)

// NewEnvironmentService returns an empty service with no active environment.
func NewEnvironmentService() *EnvironmentService {
	return &EnvironmentService{
		envs:    make(map[string]map[string]string),
		dotenv:  make(map[string]string),
		secrets: make(map[string]bool),
	}
}

// SetEnvironment replaces the variables for the named environment.
func (e *EnvironmentService) SetEnvironment(name string, vars map[string]string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	cp := make(map[string]string, len(vars))
	for k, v := range vars {
		cp[k] = v
	}
	e.envs[name] = cp
}

// SetActive selects the active environment by name.
func (e *EnvironmentService) SetActive(name string) {
	e.mu.Lock()
	e.active = name
	e.mu.Unlock()
}

// Active returns the active environment name.
func (e *EnvironmentService) Active() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.active
}

// MarkSecret flags a variable name so Mask redacts its resolved value.
func (e *EnvironmentService) MarkSecret(name string) {
	e.mu.Lock()
	e.secrets[name] = true
	e.mu.Unlock()
}

// LoadDotEnv parses a .env file (KEY=VALUE per line, # comments, optional
// quotes) and layers it beneath the active environment.
func (e *EnvironmentService) LoadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	parsed := make(map[string]string)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		parsed[k] = v
	}
	if err := sc.Err(); err != nil {
		return err
	}
	e.mu.Lock()
	e.dotenv = parsed
	e.mu.Unlock()
	return nil
}

// lookup resolves a single variable name following the precedence order.
func (e *EnvironmentService) lookup(name string) (string, bool) {
	if active, ok := e.envs[e.active]; ok {
		if v, ok := active[name]; ok {
			return v, true
		}
	}
	if v, ok := e.dotenv[name]; ok {
		return v, true
	}
	if v, ok := os.LookupEnv(name); ok {
		return v, true
	}
	return "", false
}

// Resolve expands ${VAR} / ${VAR:-default} references in s. A \${...} sequence
// is emitted literally as ${...}. A variable whose value itself contains a
// reference is expanded by recursing into that value (bounded by a depth guard
// to avoid infinite loops on self-reference) — crucially, an escaped literal is
// never re-scanned, because expansion is per-match rather than a whole-string
// fixed-point pass.
func (e *EnvironmentService) Resolve(s string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.resolve(s, 0)
}

func (e *EnvironmentService) resolve(s string, depth int) string {
	const maxDepth = 10
	return varPattern.ReplaceAllStringFunc(s, func(m string) string {
		sub := varPattern.FindStringSubmatch(m)
		esc, name, def := sub[1], sub[2], sub[3]
		if esc == `\` {
			return strings.TrimPrefix(m, `\`) // drop the escape, keep ${...} literal
		}
		v, ok := e.lookup(name)
		if !ok {
			v = def
		}
		if depth < maxDepth && strings.Contains(v, "${") {
			return e.resolve(v, depth+1)
		}
		return v
	})
}

// Mask returns s with the resolved values of secret-flagged variables replaced
// by ••••, for safe display in the UI and log scrubbing.
func (e *EnvironmentService) Mask(s string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for name := range e.secrets {
		if v, ok := e.lookup(name); ok && v != "" {
			s = strings.ReplaceAll(s, v, "••••")
		}
	}
	return s
}
