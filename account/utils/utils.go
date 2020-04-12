package utils

func StringSlice2InterfaceSlice(ss []string) []interface{} {
	is := make([]interface{}, len(ss))
	for idx, v := range ss {
		is[idx] = v
	}
	return is
}
