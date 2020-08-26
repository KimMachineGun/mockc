package basic

type Cache interface {
	Get(key string) (val interface{}, err error)
	Set(key string, val interface{}) (err error)
	Del(key string) (err error)
}
