package ar

import (
	"fmt"
	"io"
	"strings"
)

func wk(err error) error {
	if err == nil {
		return nil
	}
	if err == io.EOF {
		return err
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
