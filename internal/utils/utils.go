package utils

func IsStrSliceEqual(slis []string, tars []string) bool {
	if len(slis) != len(tars) {
		return false
	}

	for i, p := range tars {
		if slis[i] != p {
			return false
		}
	}

	return true
}

func HasStrSlicePrefix(slis []string, prefixs []string) bool {
	if len(slis) < len(prefixs) {
		return false
	}

	for i, p := range prefixs {
		if slis[i] != p {
			return false
		}
	}

	return true
}
