package slicex

func IsExist(fs []string, f string) bool {
	for i := 0; i < len(fs); i++ {
		if fs[i] == f {
			return true
		}
	}
	return false
}
