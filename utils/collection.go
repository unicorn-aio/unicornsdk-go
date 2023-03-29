package utils

/*
map[string]interface{} -> map[string]string
*/
func InterfaceToMapString(im interface{}) map[string]string {
	if im == nil {
		return nil
	}
	v, ok := im.(map[string]string)
	if ok {
		return v
	}

	vi, ok := im.(map[string]interface{})
	if !ok {
		return nil
	}

	sm := make(map[string]string)
	for k := range vi {
		if v, ok := vi[k].(string); ok {
			sm[k] = v
		}
	}
	return sm
}

/*
interface{} -> map[string]interface{}
*/
func InterfaceToMapInterface(im interface{}) map[string]interface{} {
	if im == nil {
		return nil
	}

	vi, ok := im.(map[string]interface{})
	if ok {
		return vi
	}
	pvi, ok := im.(*map[string]interface{})
	if ok {
		return *pvi
	}
	return nil
}

func InterfaceToString(i interface{}) string {
	if i == nil {
		return ""
	}

	v, ok := i.(string)
	if ok {
		return v
	}
	return ""
}

func InterfaceToArrayInterface(im interface{}) []interface{} {
	if im == nil {
		return nil
	}

	vi, ok := im.([]interface{})
	if ok {
		return vi
	}
	pvi, ok := im.(*[]interface{})
	if ok {
		return *pvi
	}
	return nil
}

func InterfaceToInt(i interface{}) int {
	if i == nil {
		return 0
	}

	v, ok := i.(int)
	if ok {
		return v
	}
	vf64, ok := i.(float64)
	if ok {
		return int(vf64)
	}
	return 0
}

func InterfaceToBool(i interface{}) bool {
	if i == nil {
		return false
	}

	v, ok := i.(bool)
	if ok {
		return v
	}
	return false
}

func InterfaceToFloat64(i interface{}) float64 {
	if i == nil {
		return 0
	}

	v, ok := i.(float64)
	if ok {
		return v
	}
	i64, ok := i.(int)
	if ok {
		return float64(i64)
	}
	return 0
}

func ArrayTToArrayInterface[T any](arr []T) []interface{} {
	iarr := make([]interface{}, len(arr))
	for idx, i := range arr {
		iarr[idx] = i
	}
	return iarr
}

/*
delete item fron slice in place, but the return should be catch too
*/
func ArrayDeleteItem[T comparable](arr []T, item T) []T {
	j := 0
	for _, v := range arr {
		if v != item {
			arr[j] = v
			j++
		}
	}
	arr = arr[:j]
	return arr
}
