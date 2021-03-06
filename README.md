[![PkgGoDev](https://pkg.go.dev/badge/github.com/KimMachineGun/mockc)](https://pkg.go.dev/github.com/KimMachineGun/mockc)
[![Go Report Card](https://goreportcard.com/badge/github.com/KimMachineGun/mockc)](https://goreportcard.com/report/github.com/KimMachineGun/mockc)
# Mockc

Mockc is a completely type-safe compile-time mock generator for Go. You can use it just by writing the mock generators with `mockc.Implement()` or using it with command like flags.

Check out [my blog post](https://blog.kimmachinegun.dev/2020/08/go-interface-mocking-with-mockc.html) for more details.

## Features

- Tools
  - [x] Generating mock with mock generators
  - [x] Generating mock with command line flags (experimental feature)
- Generated Mock
  - [x] Capturing params and results of the method
  - [x] Capturing method calls
  - [x] Injecting method body
  - [x] Customizing mock's field names with the prefix and the suffix
    - default: `prefix:"_"`, `suffix:""`
  - [x] Generating mock constructor

## Installation

```
go get github.com/KimMachineGun/mockc/cmd/mockc
```

## Look and Feel

_You can see more examples [here](https://github.com/KimMachineGun/mockc/tree/master/examples)._

### Target Interface

```go
package basic

type Cache interface {
	Get(key string) (val interface{}, err error)
	Set(key string, val interface{}) (err error)
	Del(key string) (err error)
}
```
If you want to generate mock that implements the above interface, follow the steps below.

### With Mock Generator

#### 1. Write Mock Generator

If you want to generate mock with mock generator, write a mock generator first. The mock will be generated in its generator path, and it'll be named its generator's name. You can write multiple generators in one file, and multiple mocks will be generated. The mock generator should be consisted of function calls of the `mockc` package.

```go
//+build mockc

package basic

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
mockc -destination=<output-file> -name=<mock-name> [-withConstructor] [-fieldNamePrefix=<prefix>] [-fieldNameSuffix=<suffix>] <target-interface-pattern> [<target-interface-pattern>]
Ex: mockc -destination=./example/mockc_gen.go -name=MockcCache github.com/KimMachineGun/mockc/example.Cache
```

If you want to customize the field names of the mock, pass string value to the `-fieldNamePrefix` or `-fieldNameSuffix`.

### Generated Mock

The `//go:generate` directive may vary depending on your mock generation command.

```go
// Code generated by Mockc. DO NOT EDIT.
// repo: https://github.com/KimMachineGun/mockc

//go:generate mockc
// +build !mockc

package basic

import "sync"

var _ interface {
	Cache
} = &MockcCache{}

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
```

## Inspired by 🙏

- [github.com/google/wire](https://github.com/google/wire)
- [github.com/sasha-s/goimpl](https://github.com/sasha-s/goimpl)
- [github.com/vektra/mockery](https://github.com/vektra/mockery)
