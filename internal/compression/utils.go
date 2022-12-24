package compression

import "strings"

func isGzipEncoded(encoding string) bool {
	return strings.Contains(encoding, "gzip")
}
