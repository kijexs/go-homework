package main

import (
	"fmt"
)

func twoSum(nums []int, target int) []int {
	numMap := map[int]int{}
	for i := range len(nums) {
		value := nums[i]
		needed := target - value
		index, ok := numMap[needed]
		if ok {
			return []int{index, i}
		}
		numMap[value] = i
	}
	return []int{}
}

func main() {
	res := twoSum([]int{3, 2, 4}, 6)
	fmt.Println(res)
}
