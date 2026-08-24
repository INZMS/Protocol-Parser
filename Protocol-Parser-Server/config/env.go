package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var envKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// LoadEnvFile loads KEY=VALUE pairs without overwriting variables already
// defined by the operating system. A missing file is treated as optional.
func LoadEnvFile(path string) error {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("打开配置文件%s失败: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("%s第%d行缺少等号", path, lineNumber)
		}
		key := strings.TrimSpace(parts[0])
		if !envKeyPattern.MatchString(key) {
			return fmt.Errorf("%s第%d行变量名无效: %s", path, lineNumber, key)
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		value, err := parseEnvValue(strings.TrimSpace(parts[1]))
		if err != nil {
			return fmt.Errorf("%s第%d行值无效: %w", path, lineNumber, err)
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("设置%s失败: %w", key, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取配置文件%s失败: %w", path, err)
	}
	return nil
}

func parseEnvValue(value string) (string, error) {
	if len(value) < 2 {
		return value, nil
	}
	if value[0] == '\'' && value[len(value)-1] == '\'' {
		return value[1 : len(value)-1], nil
	}
	if value[0] == '"' && value[len(value)-1] == '"' {
		return strconv.Unquote(value)
	}
	return value, nil
}
