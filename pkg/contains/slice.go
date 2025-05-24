package compare

func ContainsInt(list []int, i int) bool {
	for idx := range list {
		if list[idx] == i {
			return true
		}
	}
	return false
}

func ContainsInt32(list []int32, i int32) bool {
	for idx := range list {
		if list[idx] == i {
			return true
		}
	}
	return false
}

func ContainsInt64(list []int64, i int64) bool {
	for idx := range list {
		if list[idx] == i {
			return true
		}
	}
	return false
}

func ContainsUint32(list []uint32, i uint32) bool {
	for idx := range list {
		if list[idx] == i {
			return true
		}
	}
	return false
}

func ContainsUint64(list []uint64, i uint64) bool {
	for idx := range list {
		if list[idx] == i {
			return true
		}
	}
	return false
}

func ContainsString(list []string, str string) bool {
	for idx := range list {
		if list[idx] == str {
			return true
		}
	}
	return false
}

func AppendIgnoreNil(list []interface{}, elems ...interface{}) []interface{} {
	for idx := range elems {
		if elems[idx] != nil {
			list = append(list, elems[idx])
		}
	}
	return list
}
