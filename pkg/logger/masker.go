package logger

import (
	"regexp"
)

// MaskSensitiveData скрывает значения чувствительных полей в JSON-строке
func MaskSensitiveData(body string) string {
	re := regexp.MustCompile(`("(password|token)"\s*:\s*")([^"]*)`)
	return re.ReplaceAllString(body, `$1***`)
}
