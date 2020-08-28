package constructor

import (
	"testing"
)

func HasKey(c Cache, key string) (bool, error) {
	val, err := c.Get(key)
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func TestHasKey_WithConstructor(t *testing.T) {
	m := NewMockcCache()

	// set return value
	m._Get.Results.R0 = struct{}{}

	// execute
	key := "test_key"
	result, err := HasKey(m, key)

	// assert
	if !result {
		t.Error("result should be true")
	}
	if err != nil {
		t.Error("err should be nil")
	}
	if m._Get.CallCount != 1 {
		t.Errorf("Cache.Get should be called once: actual(%d)", m._Get.CallCount)
	}
	if m._Get.Params.P0 != key {
		t.Errorf("Cache.Get should be called with %q: actual(%q)", key, m._Get.Params.P0)
	}
}

func TestHasKey_WithMapCache(t *testing.T) {
	// set the underlying implementation by passing real implementation to constructor
	m := NewMockcCache(MapCache{})

	// execute
	key := "key"
	result, err := HasKey(m, key)

	// assert
	if result {
		t.Error("result should false")
	}
	if err != nil {
		t.Error("err should be nil")
	}
	if m._Get.CallCount != 1 {
		t.Errorf("Cache.Get should be called once: actual(%d)", m._Get.CallCount)
	}
	if m._Get.History[0].Params.P0 != key {
		t.Errorf("Cache.Get should be called with %q: actual(%q)", key, m._Get.History[0].Params.P0)
	}
}
