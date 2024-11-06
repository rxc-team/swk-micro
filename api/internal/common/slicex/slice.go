package slicex

// StringSliceEqual 比较切片数组是否相等
func StringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// StringSliceCompare 比较切片数组,返回两边不等的数据
func StringSliceCompare(a, b []string) (l, r []string) {

	if len(a) == 0 {
		return nil, b
	}
	if len(b) == 0 {
		return a, nil
	}
	var left []string
	for _, a1 := range a {
		exist := false
	ll:
		for _, b1 := range b {
			if b1 == a1 {
				exist = true
				break ll
			}
		}

		if !exist {
			left = append(left, a1)
		}
	}
	var right []string
	for _, b1 := range b {
		exist := false
	lr:
		for _, a1 := range a {
			if b1 == a1 {
				exist = true
				break lr
			}
		}

		if !exist {
			right = append(right, b1)
		}
	}

	return left, right
}
