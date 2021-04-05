// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package utils

import (
	"io"
	"sync"
)

// Ensure, that RepeaterInterfaceMock does implement RepeaterInterface.
// If this is not the case, regenerate this file with moq.
var _ RepeaterInterface = &RepeaterInterfaceMock{}

// RepeaterInterfaceMock is a mock implementation of RepeaterInterface.
//
// 	func TestSomethingThatUsesRepeaterInterface(t *testing.T) {
//
// 		// make and configure a mocked RepeaterInterface
// 		mockedRepeaterInterface := &RepeaterInterfaceMock{
// 			MakeRequestFunc: func(httpMethod Method, data io.Reader) ([]byte, error) {
// 				panic("mock out the MakeRequest method")
// 			},
// 		}
//
// 		// use mockedRepeaterInterface in code that requires RepeaterInterface
// 		// and then make assertions.
//
// 	}
type RepeaterInterfaceMock struct {
	// MakeRequestFunc mocks the MakeRequest method.
	MakeRequestFunc func(httpMethod Method, data io.Reader) ([]byte, error)

	// calls tracks calls to the methods.
	calls struct {
		// MakeRequest holds details about calls to the MakeRequest method.
		MakeRequest []struct {
			// HttpMethod is the httpMethod argument value.
			HttpMethod Method
			// Data is the data argument value.
			Data io.Reader
		}
	}
	lockMakeRequest sync.RWMutex
}

// MakeRequest calls MakeRequestFunc.
func (mock *RepeaterInterfaceMock) MakeRequest(httpMethod Method, data io.Reader) ([]byte, error) {
	if mock.MakeRequestFunc == nil {
		panic("RepeaterInterfaceMock.MakeRequestFunc: method is nil but RepeaterInterface.MakeRequest was just called")
	}
	callInfo := struct {
		HttpMethod Method
		Data       io.Reader
	}{
		HttpMethod: httpMethod,
		Data:       data,
	}
	mock.lockMakeRequest.Lock()
	mock.calls.MakeRequest = append(mock.calls.MakeRequest, callInfo)
	mock.lockMakeRequest.Unlock()
	return mock.MakeRequestFunc(httpMethod, data)
}

// MakeRequestCalls gets all the calls that were made to MakeRequest.
// Check the length with:
//     len(mockedRepeaterInterface.MakeRequestCalls())
func (mock *RepeaterInterfaceMock) MakeRequestCalls() []struct {
	HttpMethod Method
	Data       io.Reader
} {
	var calls []struct {
		HttpMethod Method
		Data       io.Reader
	}
	mock.lockMakeRequest.RLock()
	calls = mock.calls.MakeRequest
	mock.lockMakeRequest.RUnlock()
	return calls
}
