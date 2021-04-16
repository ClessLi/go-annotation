package annotation

import (
	"fmt"
	"reflect"
	"testing"
)

type TestEventBus struct {
}

//@EventBus
func (t *TestEventBus) Post(test string) {
	fmt.Println("post:", test)
}

//@EventBus
func (t *TestEventBus) RegisterInt(i int) error {
	fmt.Printf("register int: %d\n", i)
	return nil
}

func (t *TestEventBus) UnRegister(object interface{}) {
	fmt.Printf("unregister: %v\n", object)
}

type TestProxy struct {
}

func (t TestProxy) GetProxyName() string {
	return "TestProxy"
}

func (t TestProxy) Before(delegate AnnotatedMethod) bool {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()
	params := delegate.GetParams()
	fmt.Printf("before check for method %v\n", delegate.GetMethodLocation())
	if len(params) > 0 {
		for i := 0; i < len(params); i++ {
			fmt.Printf("before check: method %v, param %d, type %v, value '%v'\n", delegate.GetMethodLocation(), i, params[i].Kind(), params[i])
		}
	}
	return true
}

func (t TestProxy) After(delegate AnnotatedMethod) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()
	result := delegate.GetResult()
	fmt.Printf("after handle for method %v\n", delegate.GetMethodLocation())
	if len(result) > 0 {
		for i := 0; i < len(result); i++ {
			fmt.Printf("after handle: method %v, result %d, type %v, value '%v'\n", delegate.GetMethodLocation(), i, result[i].Kind(), result[i])
		}
	}
}

func (t TestProxy) Finally(delegate AnnotatedMethod) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()
	params := delegate.GetParams()
	result := delegate.GetResult()
	fmt.Printf("finally handle for method %v\n", delegate.GetMethodLocation())
	if len(result) > 0 {
		for i := 0; i < len(result); i++ {
			fmt.Printf("finally handle: method %v, param %d, type %v, value '%v'\n", delegate.GetMethodLocation(), i, params[i].Kind(), params[i])
			fmt.Printf("finally handle: method %v, result %d, type %v, value '%v'\n", delegate.GetMethodLocation(), i, result[i].Kind(), result[i])
		}
	}
}

func TestNewAnnotation(t *testing.T) {
	a := NewAnnotation("EventBus")
	test := new(TestEventBus)
	testProxy := new(TestProxy)
	a.RegisterAnnotatedObject(test)
	err := a.RegisterAnnotatedObjectProxy(testProxy)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", a.GetAnnotatedMethodInfos())
	msg := "test message"
	var msgI = 100
	receiverV := reflect.ValueOf(test)

	msgV := reflect.ValueOf(msg)
	msgIV := reflect.ValueOf(msgI)
	for _, methodInfo := range a.GetAnnotatedMethodInfos() {
		methodV := receiverV.MethodByName(methodInfo.MethodName)
		if methodV.Type().NumIn() != 1 {
			continue
		}
		for i := 0; i < methodV.Type().NumIn(); i++ {
			if msgV.Kind() == methodV.Type().In(i).Kind() {
				methodV.Call([]reflect.Value{msgV})
			}

			if msgIV.Kind() == methodV.Type().In(i).Kind() {
				methodV.Call([]reflect.Value{msgIV})
			}

		}
	}
}
