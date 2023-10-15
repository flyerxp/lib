package array

import (
	"strconv"
	"strings"
)

func ArrValToInt(Ids []int32, isFilter bool) []int {
	delIds := make([]int, 0, len(Ids))
	for _, v := range Ids {
		if isFilter {
			if vv := int(v); vv > 0 {
				delIds = append(delIds, vv)
			}
		} else {
			delIds = append(delIds, int(v))
		}
	}
	return delIds
}
func ArrValToString(Ids []int32, isFilter bool) []string {
	delIds := make([]string, 0, len(Ids))
	for _, v := range Ids {
		if isFilter {
			if vv := int(v); vv > 0 {
				delIds = append(delIds, strconv.Itoa(vv))
			}
		} else {
			delIds = append(delIds, strconv.Itoa(int(v)))
		}
	}
	return delIds
}
func StringToArrayString(strIds string, isFilter bool) []string {
	Ids := strings.Split(strIds, ",")
	delIds := make([]string, 0, len(Ids))
	for _, v := range Ids {
		if isFilter {
			if vv, e := strconv.Atoi(v); vv > 0 && e == nil {
				delIds = append(delIds, v)
			}
		} else {
			_, e := strconv.Atoi(v)
			if e == nil {
				delIds = append(delIds, v)
			}
		}
	}
	return delIds
}
func GetIntMapKey(m map[int]any) []int {
	t := make([]int, 0, len(m))
	for i := range m {
		t = append(t, i)
	}
	return t
}
func GetInt32MapKey(m map[int32]any) []int32 {
	t := make([]int32, 0, len(m))
	for i := range m {
		t = append(t, i)
	}
	return t
}
func GetStringMapKey(m map[string]any) []string {
	t := make([]string, 0, len(m))
	for i := range m {
		t = append(t, i)
	}
	return t
}
