// Code generated by mockery v1.0.0. DO NOT EDIT.

package bot

import mock "github.com/stretchr/testify/mock"

// MockBot is an autogenerated mock type for the Bot type
type MockBot struct {
	mock.Mock
}

// Help provides a mock function with given fields:
func (_m *MockBot) Help() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// OnMessage provides a mock function with given fields: msg
func (_m *MockBot) OnMessage(msg Message) *Response {
	ret := _m.Called(msg)

	var r0 *Response
	if rf, ok := ret.Get(0).(func(Message) *Response); ok {
		r0 = rf(msg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Response)
		}
	}

	return r0
}