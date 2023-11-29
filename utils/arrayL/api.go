package arrayL

// 去重
func UniqArr(nums []int) []int {
	unique := make(map[int]bool)
	result := []int{}
	for _, num := range nums {
		if !unique[num] {
			result = append(result, num)
			unique[num] = true
		}
	}
	return result
}
