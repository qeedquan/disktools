package ar

import (
	"fmt"
	"strings"
)

func wk(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("ar: %v", err)
}

func trim(s []byte) string {
	p := strings.TrimRight(string(s), " ")
	if strings.HasSuffix(p, "/") {
		p = strings.TrimRight(p, "/")
	}
	return p
}

func expand(p []byte, s string) {
	copy(p, s[:])
	for i := len(s); i < len(p); i++ {
		p[i] = ' '
	}
	p[len(p)-1] = '/'
}
