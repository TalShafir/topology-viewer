package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func PrefixWithKubectl(s string) string {
	if strings.HasPrefix(filepath.Base(os.Args[0]), "kubectl-") {
		return fmt.Sprintf(`kubectl %s`, s)
	}

	return s
}
