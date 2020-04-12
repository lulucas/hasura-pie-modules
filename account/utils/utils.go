package utils

func StringSlice2InterfaceSlice(ss []string) []interface{} {
	is := make([]interface{}, len(ss))
	for _, v := range ss {
		is = append(is, v)
	}
	return is
}
