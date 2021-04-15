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
func (t *TestEventBus) RegisterInt(i int) {
	fmt.Printf("register int: %d\n", i)
}

func (t *TestEventBus) UnRegister(object interface{}) {
	fmt.Printf("unregister: %v\n", object)
}

func TestNewAnnotation(t *testing.T) {
	a := NewAnnotation("EventBus")
	test := new(TestEventBus)
	a.RegisterAnnotatedObject(test)
	t.Logf("%v", a.GetAnnotatedMethodInfos())
	msg := "test message"
	var msgI uint = 100
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
