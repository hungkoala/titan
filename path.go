package titan

import (
	"path/filepath"
	"strings"
)

func AbsPathify(inPath string) string {
	if filepath.IsAbs(inPath) {
		return filepath.Clean(inPath)
	}

	p, err := filepath.Abs(inPath)
	if err == nil {
		return filepath.Clean(p)
	}

	return ""
}

//func extractTopicFromHttpUrl(path string) string {
//	if strings.HasPrefix(path, "/api/service/") || strings.HasPrefix(path, "/api/app/") {
//		return "/" + strings.Join(strings.Split(path, "/")[4:], "/")
//	}
//	return path
//}

func extractSomePartsFromUrl(url string, numberOfPart int, separator string) string {
	if url == "" {
		return url
	}
	if !strings.Contains(url, "/") {
		return url
	}
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	s := strings.Split(url, "/")
	l := numberOfPart + 1
	if len(s) < l {
		l = len(s)
	}
	return strings.Join(s[1:l], separator)
}

func ExtractLoggablePartsFromUrl(url string) string {
	return extractSomePartsFromUrl(url, 4, "/")
}

func Url2Subject(url string) string {
	return extractSomePartsFromUrl(url, 3, ".")
}

func SubjectToUrl(subject, topic string) string {
	parts := strings.Split(subject, ".")
	parts = append(parts, topic)
	return "/" + strings.Join(parts, "/")
}
