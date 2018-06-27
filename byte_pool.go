package resp

import (
	"sync"
	"math"
)

func init(){

}

func BinarySearch(v int, sorted []int) (int) {
	if sorted[0] > v {
		return 0
	}
	if sorted[len(sorted)-1] < v{
		return -1
	}
	return innerBinarySearch(v, 0, len(sorted), sorted)
}

func innerBinarySearch(v int, start int, end int, target []int) (int) {
	if start >= end{
		return end
	}
	mid := (start + end )/2
	if mid == start{
		return end
	}
	if v < pool_keys[mid] {
		return innerBinarySearch(v, start,  mid,target)
	} else if v > pool_keys[mid]{
		return innerBinarySearch(v, mid, end, target)
	} else {
		return mid
	}
}

var (
	pool_keys = []int{
		int(math.Pow(2,4)),
		int(math.Pow(2,5)),
		int(math.Pow(2,6)),
		int(math.Pow(2,7)),
		int(math.Pow(2,8)),
		int(math.Pow(2,9)),
		int(math.Pow(2,10)),
		int(math.Pow(2,11)),
		int(math.Pow(2,12)),
		int(math.Pow(2,13)),
		int(math.Pow(2,14)),
		int(math.Pow(2,15)),
		int(math.Pow(2,16)),
		int(math.Pow(2,17)),
		int(math.Pow(2,18)),
		int(math.Pow(2,19)),
		int(math.Pow(2,20)),
	}

)

type pooled_bytes_obj struct {
	size int
	pool *sync.Pool
}

type BytePoolManager struct {
	pools map[int]*pooled_bytes_obj
}

func NewBytePoolManager() *BytePoolManager{
	pool_map := make(map[int]*pooled_bytes_obj, len(pool_keys))
	for i := 4;i < 20; i ++ {
		max_capacity :=  int(math.Pow(2, float64(i)))
		pool_map [max_capacity] = &pooled_bytes_obj{
			size : max_capacity,
			pool : &sync.Pool{
				New: func() interface{} {
					return make([]byte, max_capacity)
				},
			},
		}
	}
	return &BytePoolManager{
		pools:pool_map,
	}
}

func (mgr *BytePoolManager) Get(size int) []byte {
	if size > int(math.Pow(2, 20)) {
		return make([]byte, size)
	}
	return (mgr.pools[BinarySearch(size, pool_keys)].pool.Get()).([]byte)
}

func (mgr *BytePoolManager) Put(val []byte)  {
	size := cap(val)
	index := math.Log2(float64(size))
	mgr.pools[int(index)].pool.Put(val)
}


