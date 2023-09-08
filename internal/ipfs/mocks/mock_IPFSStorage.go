// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	io "io"

	mock "github.com/stretchr/testify/mock"

	shell "github.com/ipfs/go-ipfs-api"
)

// MockIPFSStorage is an autogenerated mock type for the IPFSStorage type
type MockIPFSStorage struct {
	mock.Mock
}

type MockIPFSStorage_Expecter struct {
	mock *mock.Mock
}

func (_m *MockIPFSStorage) EXPECT() *MockIPFSStorage_Expecter {
	return &MockIPFSStorage_Expecter{mock: &_m.Mock}
}

// Add provides a mock function with given fields: payload, opts
func (_m *MockIPFSStorage) Add(payload io.Reader, opts ...func(*shell.RequestBuilder) error) (string, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, payload)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(io.Reader, ...func(*shell.RequestBuilder) error) (string, error)); ok {
		return rf(payload, opts...)
	}
	if rf, ok := ret.Get(0).(func(io.Reader, ...func(*shell.RequestBuilder) error) string); ok {
		r0 = rf(payload, opts...)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(io.Reader, ...func(*shell.RequestBuilder) error) error); ok {
		r1 = rf(payload, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockIPFSStorage_Add_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Add'
type MockIPFSStorage_Add_Call struct {
	*mock.Call
}

// Add is a helper method to define mock.On call
//   - payload io.Reader
//   - opts ...func(*shell.RequestBuilder) error
func (_e *MockIPFSStorage_Expecter) Add(payload interface{}, opts ...interface{}) *MockIPFSStorage_Add_Call {
	return &MockIPFSStorage_Add_Call{Call: _e.mock.On("Add",
		append([]interface{}{payload}, opts...)...)}
}

func (_c *MockIPFSStorage_Add_Call) Run(run func(payload io.Reader, opts ...func(*shell.RequestBuilder) error)) *MockIPFSStorage_Add_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*shell.RequestBuilder) error, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(func(*shell.RequestBuilder) error)
			}
		}
		run(args[0].(io.Reader), variadicArgs...)
	})
	return _c
}

func (_c *MockIPFSStorage_Add_Call) Return(_a0 string, _a1 error) *MockIPFSStorage_Add_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockIPFSStorage_Add_Call) RunAndReturn(run func(io.Reader, ...func(*shell.RequestBuilder) error) (string, error)) *MockIPFSStorage_Add_Call {
	_c.Call.Return(run)
	return _c
}

// Cat provides a mock function with given fields: cid
func (_m *MockIPFSStorage) Cat(cid string) (io.ReadCloser, error) {
	ret := _m.Called(cid)

	var r0 io.ReadCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (io.ReadCloser, error)); ok {
		return rf(cid)
	}
	if rf, ok := ret.Get(0).(func(string) io.ReadCloser); ok {
		r0 = rf(cid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(cid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockIPFSStorage_Cat_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Cat'
type MockIPFSStorage_Cat_Call struct {
	*mock.Call
}

// Cat is a helper method to define mock.On call
//   - cid string
func (_e *MockIPFSStorage_Expecter) Cat(cid interface{}) *MockIPFSStorage_Cat_Call {
	return &MockIPFSStorage_Cat_Call{Call: _e.mock.On("Cat", cid)}
}

func (_c *MockIPFSStorage_Cat_Call) Run(run func(cid string)) *MockIPFSStorage_Cat_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockIPFSStorage_Cat_Call) Return(_a0 io.ReadCloser, _a1 error) *MockIPFSStorage_Cat_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockIPFSStorage_Cat_Call) RunAndReturn(run func(string) (io.ReadCloser, error)) *MockIPFSStorage_Cat_Call {
	_c.Call.Return(run)
	return _c
}

// Version provides a mock function with given fields:
func (_m *MockIPFSStorage) Version() (string, string, error) {
	ret := _m.Called()

	var r0 string
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func() (string, string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() string); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func() error); ok {
		r2 = rf()
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// MockIPFSStorage_Version_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Version'
type MockIPFSStorage_Version_Call struct {
	*mock.Call
}

// Version is a helper method to define mock.On call
func (_e *MockIPFSStorage_Expecter) Version() *MockIPFSStorage_Version_Call {
	return &MockIPFSStorage_Version_Call{Call: _e.mock.On("Version")}
}

func (_c *MockIPFSStorage_Version_Call) Run(run func()) *MockIPFSStorage_Version_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockIPFSStorage_Version_Call) Return(_a0 string, _a1 string, _a2 error) *MockIPFSStorage_Version_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *MockIPFSStorage_Version_Call) RunAndReturn(run func() (string, string, error)) *MockIPFSStorage_Version_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockIPFSStorage creates a new instance of MockIPFSStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockIPFSStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockIPFSStorage {
	mock := &MockIPFSStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}