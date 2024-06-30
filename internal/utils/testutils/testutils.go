package testutils

import (
	"fmt"
	"strings"
)

// Pretty format struct
func FormatStruct(s any, msg ...string) string {
	return strings.Join(
		msg,
		" ",
	) + "\n" + strings.ReplaceAll(
		fmt.Sprintf("%+v", s),
		" ",
		"\n",
	)
}
