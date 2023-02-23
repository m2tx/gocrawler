package utils

func SliceContainsElement(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func SliceContainsSlice(s1 []string, s2 []string) bool {
	if len(s2) > len(s1) {
		return false
	}
	for _, e := range s2 {
		if !SliceContainsElement(s1, e) {
			return false
		}
	}
	return true
}
