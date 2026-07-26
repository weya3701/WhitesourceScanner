package fingerprint

import (
	"fmt"
	"regexp"
	"strings"
)

var defaultIgnorePatterns = []string{
	".git/",
	".DS_Store",
	"Thumbs.db",
	"*.swp",
	"*~",
}

type ignoreRule struct {
	raw      string
	negate   bool
	dirOnly  bool
	hasSlash bool
	re       *regexp.Regexp
}

func compileIgnoreRules(patterns []string) ([]ignoreRule, error) {
	rules := make([]ignoreRule, 0, len(patterns))
	for _, raw := range patterns {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rule := ignoreRule{raw: line}
		if strings.HasPrefix(line, `\#`) || strings.HasPrefix(line, `\!`) {
			line = line[1:]
		} else if strings.HasPrefix(line, "!") {
			rule.negate = true
			line = line[1:]
		}
		line = strings.ReplaceAll(line, `\`, "/")
		rule.dirOnly = strings.HasSuffix(line, "/")
		line = strings.TrimSuffix(line, "/")
		line = strings.TrimPrefix(line, "/")
		if line == "" {
			continue
		}
		rule.hasSlash = strings.Contains(line, "/")
		expression := globExpression(line, rule.hasSlash)
		compiled, err := regexp.Compile(expression)
		if err != nil {
			return nil, fmt.Errorf("invalid ignore pattern %q: %w", raw, err)
		}
		rule.re = compiled
		rules = append(rules, rule)
	}
	return rules, nil
}

func globExpression(pattern string, hasSlash bool) string {
	var b strings.Builder
	if hasSlash {
		b.WriteString("^")
	} else {
		b.WriteString(`(^|.*/)`)
	}
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				b.WriteString(".*")
				i += 2
			} else {
				b.WriteString(`[^/]*`)
				i++
			}
		case '?':
			b.WriteString(`[^/]`)
			i++
		case '[':
			end := strings.IndexByte(pattern[i+1:], ']')
			if end < 0 {
				b.WriteString(`\[`)
				i++
				continue
			}
			end += i + 1
			b.WriteString(pattern[i : end+1])
			i = end + 1
		default:
			b.WriteString(regexp.QuoteMeta(string(pattern[i])))
			i++
		}
	}
	b.WriteString("$")
	return b.String()
}

func ignored(path string, isDir bool, rules []ignoreRule) bool {
	result := false
	for _, rule := range rules {
		if rule.dirOnly && !isDir {
			continue
		}
		if rule.re.MatchString(path) {
			result = !rule.negate
		}
	}
	return result
}
