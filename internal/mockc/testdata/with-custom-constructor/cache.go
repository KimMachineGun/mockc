package constructor

type Cache interface {
	Get(key string) (val interface{}, err error)
	Set(key string, val interface{}) (err error)
	Del(key string) (err error)
}

type MapCache struct {
	m map[string]interface{}
}

func (c MapCache) Get(key string) (interface{}, error) {
	if c.m != nil {
		c.m = map[string]interface{}{}
	}
	return c.m[key], nil
}

func (c MapCache) Set(key string, val interface{}) error {
	if c.m != nil {
		c.m = map[string]interface{}{}
	}
	c.m[key] = val
	return nil
}

func (c MapCache) Del(key string) error {
	if c.m != nil {
		c.m = map[string]interface{}{}
	}
	delete(c.m, key)
	return nil
}
