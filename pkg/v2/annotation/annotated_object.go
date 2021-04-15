package annotation

import (
	"github.com/ClessLi/go-annotation/pkg/v2/aop"
	"reflect"
)

type AnnotatedMethod interface {
	GetMethodLocation() string
	GetMethod() reflect.Method
	GetParams() []reflect.Value
	GetResult() []reflect.Value
}

type annotatedMethod struct {
	//*sync.Mutex
	receiver       interface{}
	methodLocation string
	method         reflect.Method
	params         []reflect.Value
	result         []reflect.Value
}

func newAnnotatedMethod(delegate *aop.Delegate, methodLocation string) AnnotatedMethod {
	am := &annotatedMethod{
		receiver:       delegate.Receiver,
		methodLocation: methodLocation,
		method:         delegate.Method,
		params:         delegate.Params,
		result:         delegate.Result,
	}
	return AnnotatedMethod(am)
}

func (a annotatedMethod) GetMethodLocation() string {
	return a.methodLocation
}

func (a annotatedMethod) GetMethod() reflect.Method {
	return a.method
}

func (a annotatedMethod) GetParams() []reflect.Value {
	return a.params
}

func (a annotatedMethod) GetResult() []reflect.Value {
	return a.result
}
