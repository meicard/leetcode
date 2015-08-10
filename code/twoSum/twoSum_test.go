package twoSum

import (
	"fmt"
	"testing"
)

func TestTwoSum(t *testing.T) {
	numbers := []int{1, 3, 4, 5, 12, 34, 54, 22, 21}
	fmt.Println(twoSum(numbers, 6))
	fmt.Println(twoSum(numbers, 88))
	fmt.Println(twoSum(numbers, 89))
	fmt.Println(twoSum(numbers, 43))
	t.Log("OK")
}
