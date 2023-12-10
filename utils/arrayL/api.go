package arrayL

import "strconv"

// 去重
func UniqArr(nums []int) []int {
	unique := make(map[int]bool)
	var result []int
	for i := range nums {
		if !unique[nums[i]] {
			result = append(result, nums[i])
			unique[nums[i]] = true
		}
	}
	return result
}
func UniqArr32(nums []int32) []int32 {
	unique := make(map[int32]bool)
	var result []int32
	for i := range nums {
		if !unique[nums[i]] {
			result = append(result, nums[i])
			unique[nums[i]] = true
		}
	}
	return result
}
func UniqArr32ToInt(nums []int32) []int {
	unique := make(map[int32]bool)
	var result []int
	for i := range nums {
		if !unique[nums[i]] {
			result = append(result, int(nums[i]))
			unique[nums[i]] = true
		}
	}
	return result
}
func UniqArr32ToString(nums []int32) []string {
	unique := make(map[int32]bool)
	var result []string
	for i := range nums {
		if !unique[nums[i]] {
			result = append(result, strconv.Itoa(int(nums[i])))
			unique[nums[i]] = true
		}
	}
	return result
}
func UniqArrToString(nums []int) []string {
	unique := make(map[int]bool)
	var result []string
	for i := range nums {
		if !unique[nums[i]] {
			result = append(result, strconv.Itoa(nums[i]))
			unique[nums[i]] = true
		}
	}
	return result
}
