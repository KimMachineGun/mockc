[![PkgGoDev](https://pkg.go.dev/badge/github.com/KimMachineGun/mockc)](https://pkg.go.dev/github.com/KimMachineGun/mockc)
[![Go Report Card](https://goreportcard.com/badge/github.com/KimMachineGun/mockc)](https://goreportcard.com/report/github.com/KimMachineGun/mockc)
# Mockc: Complie-time mock generator for Go
Mockc is a completely type-safe compile-time mock generator for Go. You can use it just by writing the mock generators with `mockc.Implement()` or using it with command like flags.

## Features
- Tools
  - [x] Generating mock with mock generators
  - [x] Generating mock with command line flags (experimental feature)
- Generated Mock
  - [x] Capturing params and results of the method
  - [x] Capturing method calls
  - [x] Injecting method body
  - [x] Customizing mock's field names with prefix and suffix
    - default: `prefix:"_"`, `suffix:""`

## Installation
```
go get github.com/KimMachineGun/mockc/cmd/mockc
```

## Look and Feel
### Target Interface
```go
package ex

type Cache interface {
	Get(key string) (val interface{}, err error)
	Set(key string, val interface{}) (err error)
	Del(key string) (err error)
}
```
If you want to generate mock that implements the above interface, follow the steps below.
### With Mock Generator
#### 1. Write Mock Generator
If you want to generate mock with mock generator, write a mock generator first. The mock will be generated in its generator path and it'll be named its generator's name. You can write multiple generators in one file, and multiple mocks will be generated. The mock generator should be consist of function calls of the `mockc` package.
```go
//+build mockc

package ex

import (
	"github.com/KimMachineGun/mockc"
)

func MockcCache() {
	mockc.Implement(Cache(nil))
}
```
If you want to customize the field names of the mock, use `mockc.SetFieldNamePrefix()` or `mockc.SetFieldNameSuffix()`. (Notice: These functions only work with constant string value.)
#### 2. Generate Mock
This command will generate mock with your mock generator. The `<package-pattern>` argument will be used for loading mock generator with [golang.org/x/tools/go/packages#Load](https://godoc.org/golang.org/x/tools/go/packages#Load). If it's not provided, `.` will be used.
```sh
mockc [<packages-pattern>]
Ex: mock ./example
```
### With Command Line Flags
#### 1. Generate Mock
This command will generate mock with its command line flags. If you generate mock with this command, you don't need to write the mock generator. The `<target-interface-pattern>` should follow `{package_path}.{interface_name}` format.
```sh
mockc -name=<mock-name> -destination=<output-file> [-fieldNamePrefix=<prefix>] [-fieldNameSuffix=<suffix>] <target-interface-pattern> [<target-interface-pattern>]
Ex: mockc -name=MockcCache -destination=./example/mockc_gen.go github.com/KimMachineGun/mockc/example.Cache
```
If you want to customize the field names of the mock, pass string value to the `-fieldNamePrefix` or `-fieldNameSuffix`.
### Generated Mock	
The `//go:generate` comment may vary depending on your mock generation command.
```go
// Code generated by Mockc. DO NOT EDIT.
// repo: https://github.com/KimMachineGun/mockc

//go:generate mockc
// +build !mockc

package ex

import (
	"sync"
)

type MockcCache struct {
	// method: Del
	_Del struct {
		mu sync.Mutex
		// basics
		Called    bool
		CallCount int
		// call history
		History []struct {
			Params struct {
				P0 string
			}
			Results struct {
				R0 error
			}
		}
		// params
		Params struct {
			P0 string
		}
		// results
		Results struct {
			R0 error
		}
		// if it is not nil, it'll be called in the middle of the method.
		Body func(string) error
	}
	// method: Get
	_Get struct {
		mu sync.Mutex
		// basics
		Called    bool
		CallCount int
		// call history
		History []struct {
			Params struct {
				P0 string
			}
			Results struct {
				R0 interface{}
				R1 error
			}
		}
		// params
		Params struct {
			P0 string
		}
		// results
		Results struct {
			R0 interface{}
			R1 error
		}
		// if it is not nil, it'll be called in the middle of the method.
		Body func(string) (interface{}, error)
	}
	// method: Set
	_Set struct {
		mu sync.Mutex
		// basics
		Called    bool
		CallCount int
		// call history
		History []struct {
			Params struct {
				P0 string
				P1 interface{}
			}
			Results struct {
				R0 error
			}
		}
		// params
		Params struct {
			P0 string
			P1 interface{}
		}
		// results
		Results struct {
			R0 error
		}
		// if it is not nil, it'll be called in the middle of the method.
		Body func(string, interface{}) error
	}
}

func (recv *MockcCache) Del(p0 string) error {
	recv._Del.mu.Lock()
	defer recv._Del.mu.Unlock()
	// basics
	recv._Del.Called = true
	recv._Del.CallCount++
	// params
	recv._Del.Params.P0 = p0
	// body
	if recv._Del.Body != nil {
		recv._Del.Results.R0 = recv._Del.Body(p0)
	}
	// call history
	recv._Del.History = append(recv._Del.History, struct {
		Params struct {
			P0 string
		}
		Results struct {
			R0 error
		}
	}{
		Params:  recv._Del.Params,
		Results: recv._Del.Results,
	})
	// results
	return recv._Del.Results.R0
}

func (recv *MockcCache) Get(p0 string) (interface{}, error) {
	recv._Get.mu.Lock()
	defer recv._Get.mu.Unlock()
	// basics
	recv._Get.Called = true
	recv._Get.CallCount++
	// params
	recv._Get.Params.P0 = p0
	// body
	if recv._Get.Body != nil {
		recv._Get.Results.R0, recv._Get.Results.R1 = recv._Get.Body(p0)
	}
	// call history
	recv._Get.History = append(recv._Get.History, struct {
		Params struct {
			P0 string
		}
		Results struct {
			R0 interface{}
			R1 error
		}
	}{
		Params:  recv._Get.Params,
		Results: recv._Get.Results,
	})
	// results
	return recv._Get.Results.R0, recv._Get.Results.R1
}

func (recv *MockcCache) Set(p0 string, p1 interface{}) error {
	recv._Set.mu.Lock()
	defer recv._Set.mu.Unlock()
	// basics
	recv._Set.Called = true
	recv._Set.CallCount++
	// params
	recv._Set.Params.P0 = p0
	recv._Set.Params.P1 = p1
	// body
	if recv._Set.Body != nil {
		recv._Set.Results.R0 = recv._Set.Body(p0, p1)
	}
	// call history
	recv._Set.History = append(recv._Set.History, struct {
		Params struct {
			P0 string
			P1 interface{}
		}
		Results struct {
			R0 error
		}
	}{
		Params:  recv._Set.Params,
		Results: recv._Set.Results,
	})
	// results
	return recv._Set.Results.R0
}
```
### Feel Free to Use the Generated Mock
```go
package ex

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
```

## Inspired by 🙏
- [github.com/google/wire](https://github.com/google/wire)
- [github.com/sasha-s/goimpl](https://github.com/sasha-s/goimpl)
- [github.com/vektra/mockery](https://github.com/vektra/mockery)
