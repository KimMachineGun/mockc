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
	mc := MapCache{}
	m := NewMockcCache(mc)

	key := "key"
	expected := false
	actual, err := HasKey(m, key)

	if actual != expected {
		t.Errorf("result: expected(%v) != actual(%v)", expected, actual)
	}
	if err != nil {
		t.Errorf("err: %v", err)
	}
	if m._Get.CallCount != 1 {
		t.Errorf("Cache.Get should be called once: actual(%d)", m._Get.CallCount)
	}
	if m._Get.History[0].Params.P0 != key {
		t.Errorf("Cache.Get should be called with %q: actual(%q)", key, m._Get.History[0].Params.P0)
	}
	if res := m._Get.History[0].Results; res.R0 != nil || res.R1 != nil {
		t.Errorf("Cache.Get should return nil, nil: actual(%v, %v)", res.R0, res.R1)
	}
}
