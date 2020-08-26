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

	m._Get.Results.R0 = struct{}{}

	key := "key"
	expected := true
	actual, err := HasKey(m, key)

	if actual != expected {
		t.Errorf("result: expected(%v) != actual(%v)", expected, actual)
	}
	if err != nil {
		t.Errorf("err: %v", err)
	}
	if m._Get.Params.P0 != key {
		t.Errorf("expected(%v) != actual(%v)", key, m._Get.Params.P0)
	}
}

func TestHasKey_WithMethodBodyInjection(t *testing.T) {
	m := &MockcCache{}
	m._Get.Body = func(key string) (interface{}, error) {
		if key == "key" {
			return nil, errors.New("err")
		}
		return nil, nil
	}

	key := "key"
	expected := false
	actual, err := HasKey(m, key)

	if expected != actual {
		t.Errorf("result: expected(%v) != actual(%v)", expected, actual)
	}
	if err == nil {
		t.Errorf("err: %v", err)
	}
	if key != m._Get.Params.P0 {
		t.Errorf("param key: expected(%v) != actual(%v)", key, m._Get.Params.P0)
	}
}

func TestHasKey_WithHistory(t *testing.T) {
	m := &MockcCache{}

	table := []struct {
		key string
		val interface{}

		expected bool
		err      error
	}{
		{
			key: "key1",
			val: struct{}{},

			expected: true,
			err:      nil,
		},
		{

			key: "key2",
			val: nil,

			expected: false,
			err:      errors.New("err"),
		},
	}

	for _, t := range table {
		m._Get.Results.R0 = t.val
		m._Get.Results.R1 = t.err

		HasKey(m, t.key)
	}

	for idx, h := range m._Get.History {
		if expected, actual := table[idx].expected, h.Results.R0 != nil; expected != actual {
			t.Errorf("table[%v] result : expected(%v) != actual(%v)", idx, expected, actual)
		}
		if expected, actual := table[idx].err, h.Results.R1; expected != actual {
			t.Errorf("table[%v] err : expected(%v) != actual(%v)", idx, expected, actual)
		}
		if expected, actual := table[idx].key, h.Params.P0; expected != actual {
			t.Errorf("table[%v] param key: expected(%v) != actual(%v)", idx, expected, actual)
		}
	}
}
