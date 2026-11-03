package repository

import "unicode/utf8"

// clipLine укорачивает строку результата поиска, не разрывая символ
func clipLine(line string) string {
	if len(line) <= maxLineBytes {
		return line
	}
	end := maxLineBytes
	for end > 0 && !utf8.RuneStart(line[end]) {
		end--
	}
	return line[:end] + "…"
}
