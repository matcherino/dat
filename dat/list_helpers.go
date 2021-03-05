package dat

import (
	"github.com/matcherino/dat/common"
)

var bufPool = common.NewBufferPool()

func writeIdentifiers(buf common.BufferWriter, columns []string, delimiter string) {
	for i, column := range columns {
		if i > 0 {
			buf.WriteString(delimiter)
		}
		buf.WriteString(column)
	}
}

func writeIdentifier(buf common.BufferWriter, name string) {
	buf.WriteString(name)
}

func writeQuotedIdentifier(buf common.BufferWriter, name string) {
	buf.WriteRune('"')
	buf.WriteString(name)
	buf.WriteRune('"')
}

func buildPlaceholders(buf common.BufferWriter, start, length int) {
	// Build the placeholder like "($1,$2,$3)"
	buf.WriteRune('(')
	for i := start; i < start+length; i++ {
		if i > start {
			buf.WriteRune(',')
		}
		writePlaceholder(buf, i)
	}
	buf.WriteRune(')')
}

// joinPlaceholders returns $1, $2 ... , $n
func writePlaceholders(buf common.BufferWriter, length int, join string, offset int) {
	for i := 0; i < length; i++ {
		if i > 0 {
			buf.WriteString(join)
		}
		writePlaceholder(buf, i+offset)
	}
}
