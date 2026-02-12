package classify

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Category string

const (
	CategoryNormal  Category = "normal"
	CategoryScanner Category = "scanner"
	CategoryHoney   Category = "honey"
	CategoryDenied  Category = "denied"
)

type Rule struct {
	Name           string   `yaml:"name"`
	UASubstrings   []string `yaml:"ua_substrings"`
	PathSubstrings []string `yaml:"path_substrings"`
	Methods        []string `yaml:"methods"`
	Score          int      `yaml:"score"`
}

type Library struct {
	Rules []Rule `yaml:"rules"`
}

type Classifier struct {
	honeyPaths []string
	lib        Library
}

type Result struct {
	Category    Category
	Scanner     bool
	Honey       bool
	ScoreDelta  int
	RuleMatches []string
}

func New(honeyPaths []string, externalLibFile string) (*Classifier, error) {
	lib := builtInLibrary()
	if strings.TrimSpace(externalLibFile) != "" {
		ext, err := loadLibraryFile(externalLibFile)
		if err != nil {
			return nil, err
		}
		lib.Rules = append(lib.Rules, ext.Rules...)
	}
	return &Classifier{honeyPaths: honeyPaths, lib: lib}, nil
}

func (c *Classifier) Classify(r *http.Request, denied bool) Result {
	if denied {
		return Result{Category: CategoryDenied, ScoreDelta: 3}
	}
	path := strings.ToLower(r.URL.Path)
	ua := strings.ToLower(r.UserAgent())
	method := strings.ToUpper(r.Method)

	honey := matchHoney(path, c.honeyPaths)
	matchedRules := make([]string, 0)
	score := 0
	for _, rule := range c.lib.Rules {
		if !ruleMatch(rule, ua, path, method) {
			continue
		}
		score += max(rule.Score, 1)
		matchedRules = append(matchedRules, rule.Name)
	}

	scanner := score >= 2
	category := CategoryNormal
	switch {
	case honey:
		category = CategoryHoney
	case scanner:
		category = CategoryScanner
	}

	return Result{
		Category:    category,
		Scanner:     scanner,
		Honey:       honey,
		ScoreDelta:  score,
		RuleMatches: matchedRules,
	}
}

func matchHoney(path string, patterns []string) bool {
	for _, p := range patterns {
		pp := strings.ToLower(strings.TrimSpace(p))
		if pp == "" {
			continue
		}
		switch {
		case strings.HasPrefix(pp, "*") && strings.HasSuffix(pp, "*") && len(pp) > 2:
			if strings.Contains(path, pp[1:len(pp)-1]) {
				return true
			}
		case strings.Contains(pp, "*"):
			// Treat wildcards simply as contains with wildcard stripped.
			s := strings.ReplaceAll(pp, "*", "")
			if s != "" && strings.Contains(path, s) {
				return true
			}
		case strings.HasSuffix(pp, "/"):
			if strings.HasPrefix(path, pp) {
				return true
			}
		default:
			if path == pp {
				return true
			}
		}
	}
	return false
}

func ruleMatch(rule Rule, ua, path, method string) bool {
	if len(rule.Methods) > 0 {
		ok := false
		for _, m := range rule.Methods {
			if strings.EqualFold(m, method) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if len(rule.UASubstrings) > 0 {
		hit := false
		for _, s := range rule.UASubstrings {
			if strings.Contains(ua, strings.ToLower(s)) {
				hit = true
				break
			}
		}
		if !hit {
			return false
		}
	}
	if len(rule.PathSubstrings) > 0 {
		hit := false
		for _, s := range rule.PathSubstrings {
			v := strings.ToLower(s)
			if strings.HasPrefix(v, "re:") {
				rx, err := regexp.Compile(v[3:])
				if err == nil && rx.MatchString(path) {
					hit = true
					break
				}
				continue
			}
			if strings.Contains(path, v) {
				hit = true
				break
			}
		}
		if !hit {
			return false
		}
	}
	return true
}

func loadLibraryFile(path string) (Library, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Library{}, fmt.Errorf("read fingerprint library: %w", err)
	}
	var lib Library
	if err := yaml.Unmarshal(b, &lib); err != nil {
		return Library{}, fmt.Errorf("parse fingerprint library: %w", err)
	}
	return lib, nil
}

func builtInLibrary() Library {
	return Library{Rules: []Rule{
		{Name: "nmap", UASubstrings: []string{"nmap"}, Score: 3},
		{Name: "masscan", UASubstrings: []string{"masscan"}, Score: 3},
		{Name: "zgrab", UASubstrings: []string{"zgrab"}, Score: 3},
		{Name: "nikto", UASubstrings: []string{"nikto"}, Score: 3},
		{Name: "sqlmap", UASubstrings: []string{"sqlmap"}, Score: 3},
		{Name: "cms-probe", PathSubstrings: []string{"/wp-login.php", "/wp-admin", "/phpmyadmin", "/.env", "/.git"}, Score: 2},
		{Name: "method-anomaly", Methods: []string{"OPTIONS", "TRACE", "CONNECT"}, Score: 2},
	}}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
