package twoSum

func twoSum(numbers []int, target int) [2]int {
	m := make(map[int]int)
	var r [2]int
	for k, _ := range numbers {
		if _, ok := m[numbers[k]]; !ok {
			m[target-numbers[k]] = k
		} else {
			r[0] = m[numbers[k]] + 1
			r[1] = k + 1
			break
		}
	}
	return r
}
