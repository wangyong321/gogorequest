package hash

import (
	"crypto/sha1"
	"fmt"
	"io"
)

func StringToSha1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}
