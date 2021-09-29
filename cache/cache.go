package cache

// Cache 缓存接口
type Cache interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	Del(key string)
	DelOldest()
	Len() int
}



