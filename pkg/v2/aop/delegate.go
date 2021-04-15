package aop

import "reflect"

type Delegate struct {
	Receiver interface{}
	Method   reflect.Method
	Params   []reflect.Value
	Result   []reflect.Value
}

func NewDelegate(receiver interface{}, method reflect.Method, params []reflect.Value) *Delegate {
	delegate := &Delegate{
		Receiver: receiver,
		Method:   method,
		Params:   params,
	}

	fn := method.Func
	fnType := fn.Type()
	nout := fnType.NumOut()
	delegate.Result = make([]reflect.Value, nout)
	for i := 0; i < nout; i++ {
		delegate.Result[i] = reflect.Zero(fnType.Out(i))
	}

	return delegate
}
