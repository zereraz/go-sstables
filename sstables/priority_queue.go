package sstables

import (
	"github.com/thomasjungblut/go-sstables/skiplist"
)

type PriorityQueueI interface {
	Init(iterators []SSTableIteratorI) error // initializes the heap with the initial values from the iterators
	Next() ([]byte, []byte, error)           // next key/value/error, Done is returned when all elements are exhausted
}

type Element struct {
	key       []byte
	value     []byte
	heapIndex int
	iterator  SSTableIteratorI
}

type PriorityQueue struct {
	size int
	heap []*Element
	comp skiplist.KeyComparator
}

func NewPriorityQueue(comp skiplist.KeyComparator) PriorityQueue {
	return PriorityQueue{comp: comp}
}

func (pq PriorityQueue) lessThan(i, j *Element) bool {
	return pq.comp(i.key, j.key) < 0
}

func (pq PriorityQueue) swap(i, j int) {
	pq.heap[i], pq.heap[j] = pq.heap[j], pq.heap[i]
	pq.heap[i].heapIndex = i
	pq.heap[j].heapIndex = j
}

func (pq *PriorityQueue) Init(iterators []SSTableIteratorI) error {
	// reserve the 0th element for nil, makes it easier to implement the rest of the logic
	pq.heap = []*Element{nil}
	for i, it := range iterators {
		e := &Element{heapIndex: i, iterator: it, key: nil, value: nil}
		err := fillNext(e)
		if err == nil {
			pq.heap = append(pq.heap, e)
			pq.size++
			pq.upHeap(pq.size)
		} else if err != Done {
			return err
		}
	}

	return nil
}

func (pq *PriorityQueue) Next() ([]byte, []byte, error) {
	if pq.size == 0 {
		return nil, nil, Done
	}
	// since we reserved index 0 for nil, the minimum element is always at index 1
	top := pq.heap[1]
	k := top.key
	v := top.value
	err := fillNext(top)
	// if we encounter a real error, we're returning immediately
	if err != nil && err != Done {
		return nil, nil, err
	}

	// remove the element from the heap completely if its iterator is exhausted
	if err == Done {
		// move the root away to the bottom leaf
		pq.swap(1, pq.size)
		// and chop it off the slice
		pq.heap = pq.heap[0:pq.size]
		pq.size--
	}

	// always down the heap at the end
	pq.downHeap()

	return k, v, nil
}

func (pq *PriorityQueue) upHeap(i int) {
	element := pq.heap[i]
	j := i >> 1
	for j > 0 && pq.lessThan(element, pq.heap[j]) {
		pq.heap[i] = pq.heap[j]
		i = j
		j = j >> 1
	}
	pq.heap[i] = element
}

func (pq *PriorityQueue) downHeap() {
	if pq.size == 0 {
		return
	}

	i := 1
	element := pq.heap[i]
	j := i << 1
	k := j + 1
	if k <= pq.size && pq.lessThan(pq.heap[k], pq.heap[j]) {
		j = k
	}
	for j <= pq.size && pq.lessThan(pq.heap[j], element) {
		pq.heap[i] = pq.heap[j]
		i = j
		j = i << 1
		k = j + 1
		if k <= pq.size && pq.lessThan(pq.heap[k], pq.heap[j]) {
			j = k
		}
	}
	pq.heap[i] = element
}

func fillNext(item *Element) error {
	k, v, e := item.iterator.Next()
	if e != nil {
		item.key = nil
		item.value = nil
		return e
	}

	item.key = k
	item.value = v

	return nil
}
