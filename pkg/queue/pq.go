package queue

type PriorityItem interface {
	WeightAble
	IndexAble
}

type WeightAble interface {
	Weight() float64
}

type IndexAble interface {
	Index() int
	SetIndex(index int)
}

// PriorityQueue 优先级队列
// 由 slice 驱动
type PriorityQueue[T PriorityItem] struct {
	items []T
}

func NewPriorityQueue[T PriorityItem](list []T) *PriorityQueue[T] {
	return &PriorityQueue[T]{
		items: list,
	}
}

// Len 获取堆的长度
func (pq *PriorityQueue[T]) Len() int {
	len := len(pq.items)
	return len
}

// Less 判断元素在堆中的位置（用于实现Heap接口）
func (pq *PriorityQueue[T]) Less(i, j int) bool {
	return pq.items[i].Weight() > pq.items[j].Weight()
}

// Swap 用于交换元素在堆中的位置（用于实现Heap接口）
func (pq *PriorityQueue[T]) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// 将元素推入堆中（用于实现Heap接口）
func (pq *PriorityQueue[T]) Push(t any) {
	item := t.(T)
	item.SetIndex(pq.Len())
	pq.items = append(pq.items, item)
}

// 将元素从对堆中弹出（用于实现Heap接口）
func (pq *PriorityQueue[T]) Pop() any {

	lenght := pq.Len()
	old := pq.items
	newsItem := old[lenght-1]
	// 帮助垃圾回收
	var t T
	old[lenght-1] = t

	pq.items = old[0 : lenght-1]

	newsItem.SetIndex(-1)
	return newsItem
}
