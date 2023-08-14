package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	data := []int{1, 2, 3, 4}
	result := Map(data, func(x int) int {
		return x + 1
	})
	assert.ElementsMatch(t, result, []int{2, 3, 4, 5})
}

func TestReduce(t *testing.T) {

	add := func(x int, y int) int {
		return x + y
	}
  
  data := []int{}
  result := Reduce(data, add)
  assert.Equal(t, result, 0)

	data = []int{1}
	result = Reduce(data, add)
  assert.Equal(t, result, 1)

	data = []int{1, 2, 3, 4}
	result = Reduce(data, add)
  assert.Equal(t, result, 10)
}

func TestFilter(t *testing.T) {
  cmp := func(x int) bool {
    return x <= 5
  }
  
  data := []int{1,2,3,4,5,6,7,8,9}
  result := Filter(data, cmp)
  assert.ElementsMatch(t, result, []int{1,2,3,4,5})
}
