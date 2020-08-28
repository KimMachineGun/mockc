package basic

import (
	"errors"
	"testing"
)

func HasKey(c Cache, key string) (bool, error) {
	val, err := c.Get(key)
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func TestHasKey(t *testing.T) {
	m := &MockcCache{}

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

func TestHasKey_WithBodyInjection(t *testing.T) {
	m := &MockcCache{}

	// inject body
	key := "test_key"
	m._Get.Body = func(actualKey string) (interface{}, error) {
		if actualKey != key {
			t.Errorf("Cache.Get should be called with %q: actual(%q)", key, actualKey)
		}
		return nil, errors.New("error")
	}

	// execute
	result, err := HasKey(m, key)

	// assert
	if result {
		t.Error("result should be false")
	}
	if err == nil {
		t.Error("err should not be nil")
	}
	if m._Get.CallCount != 1 {
		t.Errorf("Cache.Get should be called once: actual(%d)", m._Get.CallCount)
	}
}
