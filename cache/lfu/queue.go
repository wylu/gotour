package lfu

import (
	"cache"
	"container/heap"
)

type entry struct {
	key    string
	value  interface{}
	weight int
	index  int
}

func (e *entry) Len() int {
	return cache.CalcLen(e.value)
}

type queue []*entry

func (q queue) Len() int {
	return len(q)
}

func (q queue) Less(i, j int) bool {
	if q[i].weight == q[j].weight {
		return q[i].key < q[j].key
	}
	return q[i].weight < q[j].weight
}

func (q queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *queue) Push(x interface{}) {
	et := x.(*entry)
	et.index = len(*q)
	*q = append(*q, et)
}

func (q *queue) Pop() interface{} {
	old := *q
	n := len(old)
	et := old[n-1]
	old[n-1] = nil // avoid memory leak
	et.index = -1  // for safety
	*q = old[0 : n-1]
	return et
}

// update modifies the weight and value of an entry in the queue.
func (q *queue) update(et *entry, value interface{}, weight int) {
	et.value = value
	et.weight = weight
	heap.Fix(q, et.index)
}
