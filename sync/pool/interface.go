package pool

type Pool[T any] interface {
	Get() T
	Put(T)
}

type Pools[T any] interface {
	Get(size int) T
	Put(T)
}
