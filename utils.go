package quartz

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const (
	DEFAULT_GROUP = "DEFAULT"
)

func newUniqueName(group string) string {
	buf := make([]byte, 16)

	rand.Read(buf)

	hash := md5.Sum([]byte(group))

	return fmt.Sprintf("%s-%s", hex.EncodeToString(hash[12:]), hex.EncodeToString(buf))
}
