package sstables

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"sort"
)

func IntComp(a interface{}, b interface{}) int {
	aInt := a.(int)
	bInt := b.(int)

	if aInt > bInt {
		return 1
	} else if aInt < bInt {
		return -1
	}

	return 0
}

func TestSkipListSingleInsertHappyPath(t *testing.T) {
	list := NewSkipList(IntComp)
	list.Insert(13)

	assert.Equal(t, 1, list.Size())
	assert.True(t, list.Contains(13))
	assert.False(t, list.Contains(1))

	// manually test the iterator
	it := list.Iterator()
	e, err := it.Next()
	assert.Nil(t, err)
	assert.Equal(t, 13, e.(int))
	e, err = it.Next()
	assert.Nil(t, e)
	assert.Equal(t, Done, err)
}

func TestSkipListMultiInsertOrdered(t *testing.T) {
	list := NewSkipList(IntComp)
	batchInsertAndAssertContains(t, []int{1, 2, 3, 4, 5, 6, 7}, &list)
}

func TestSkipListMultiInsertUnordered(t *testing.T) {
	list := NewSkipList(IntComp)
	batchInsertAndAssertContains(t, []int{79, 14, 91, 27, 62, 41, 58, 2, 20, 87, 34}, &list)
}

func TestSkipListMultiInsertUnorderedNegatives(t *testing.T) {
	list := NewSkipList(IntComp)
	batchInsertAndAssertContains(t, []int{79, 14, -91, 27, 62, 41, -58, 2, -20, -87, 34}, &list)
}

func TestSkipListMultiInsertZeroRun(t *testing.T) {
	list := NewSkipList(IntComp)
	batchInsertAndAssertContains(t, []int{2, 1, 0, -1, -2}, &list)
}

func TestSkipListDoubleEqualInsert(t *testing.T) {
	assert.PanicsWithValue(t, "duplicate key insertions are not allowed", func() {
		list := NewSkipList(IntComp)
		list.Insert(13)
		list.Insert(13) // should panic on duped key
	})
}

func TestSkipListEmptyIterator(t *testing.T) {
	list := NewSkipList(IntComp)

	assert.Equal(t, 0, list.Size())
	assert.False(t, list.Contains(1))

	// manually test the iterator
	it := list.Iterator()
	e, err := it.Next()
	assert.Nil(t, e)
	assert.Equal(t, Done, err)
}

func TestSkipListMultiInsertUnorderedStartingIterator(t *testing.T) {
	list := NewSkipList(IntComp)
	batchInsertAndAssertContains(t, []int{79, 14, 91, 27, 62, 41, 58, 2, 20, 87, 34}, &list)
	expected := []int{2, 14, 20, 27, 34, 41, 58, 62, 79, 87, 91}
	// a lower key of the sequence should yield the whole sequence
	it := list.IteratorStartingAt(1)
	assertIteratorOutputs(t, expected, it)

	// first key should also yield the whole sequence
	it = list.IteratorStartingAt(2)
	assertIteratorOutputs(t, expected, it)

	// test a staggered range at each index
	for i, start := range expected {
		sliced := expected[i:]
		it = list.IteratorStartingAt(start)
		assertIteratorOutputs(t, sliced, it)
	}

	// test out of range iteration, which should yield an empty iterator
	it = list.IteratorStartingAt(100)
	e, err := it.Next()
	assert.Nil(t, e)
	assert.Equal(t, Done, err)
}

func assertIteratorOutputs(t *testing.T, expectedSeq []int, it *SkipListIterator) {
	currentIndex := 0
	for {
		e, err := it.Next()
		if err == Done {
			break
		}

		if err != nil {
			assert.Fail(t, "received an error while iterating, shouldn't happen")
		}

		assert.NotNil(t, e)

		assert.Equal(t, expectedSeq[currentIndex], e.(int))
		currentIndex++
	}

}

func batchInsertAndAssertContains(t *testing.T, toInsert []int, list *SkipList) {
	for _, e := range toInsert {
		list.Insert(e)
	}
	assert.Equal(t, len(toInsert), list.Size())
	for _, e := range toInsert {
		assert.True(t, list.Contains(e))
	}

	sort.Ints(toInsert)
	it := list.Iterator()
	assertIteratorOutputs(t, toInsert, it)
}