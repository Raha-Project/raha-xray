package common

func NonEmptyValue(entry string) bool {
	switch entry {
	case "", "{}", "[]":
		return false
	}
	return true
}
