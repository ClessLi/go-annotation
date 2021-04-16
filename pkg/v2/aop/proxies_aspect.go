package aop

import (
	"bou.ke/monkey"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type ProxiesAspect interface {
	RegisterDelegate(delegateType reflect.Type)
	RegisterProxy(proxy Proxy) error
	beforeProcessed(point *Delegate, methodLocation string) bool
	afterProcessed(point *Delegate, methodLocation string)
	finallyProcessed(point *Delegate, methodLocation string)
}

type aspect struct {
	sync.RWMutex
	hasProxy  map[string]bool
	units     map[string]Proxy
	unitNames []string
}

var (
	SingletonAspect ProxiesAspect
	onceForAspect   = new(sync.Once)
)

func GetSingletonAspectInstance() ProxiesAspect {
	onceForAspect.Do(func() {
		a := &aspect{
			RWMutex:   sync.RWMutex{},
			hasProxy:  make(map[string]bool),
			units:     make(map[string]Proxy),
			unitNames: make([]string, 0),
		}
		SingletonAspect = ProxiesAspect(a)
	})
	return SingletonAspect
}

func (a *aspect) RegisterDelegate(delegateType reflect.Type) {
	a.Lock()
	defer a.Unlock()
	pkgPth := delegateType.PkgPath()
	receiverName := delegateType.Name()
	if delegateType.Kind() == reflect.Ptr {
		pkgPth = delegateType.Elem().PkgPath()
		receiverName = delegateType.Elem().Name()
	}
	for i := 0; i < delegateType.NumMethod(); i++ {
		method := delegateType.Method(i)
		pkgList := strings.Split(pkgPth, "/")
		methodLocation := fmt.Sprintf("%s.%s.%s", pkgList[len(pkgList)-1], receiverName, method.Name)
		//methodLocation := fmt.Sprintf("<%s>%s.%s", pkgPth, receiverName, method.Name)
		if has, isIn := a.hasProxy[methodLocation]; isIn && has {
			continue
		}
		//guard := new(monkey.PatchGuard)
		//var proxyProcessed = func(in []reflect.Value) []reflect.Value {
		//	guard.Unpatch()
		//	defer guard.Restore()
		//	receiver := in[0]
		//	delegate := NewDelegate(receiver, method, in[1:])
		//	defer a.finallyProcessed(delegate, methodLocation)
		//	if !a.beforeProcessed(delegate, methodLocation) {
		//		return delegate.Result
		//	}
		//	delegate.Result = receiver.MethodByName(method.Name).Call(in[1:])
		//	a.afterProcessed(delegate, methodLocation)
		//	return delegate.Result
		//}
		//proxyFn := reflect.MakeFunc(method.Func.Type(), proxyProcessed)
		//*guard = *monkey.PatchInstanceMethod(delegateType, method.Name, proxyFn.Interface())
		var guard *monkey.PatchGuard
		guard = monkey.PatchInstanceMethod(delegateType, method.Name, reflect.MakeFunc(method.Func.Type(), func(in []reflect.Value) (results []reflect.Value) {
			guard.Unpatch()
			defer guard.Restore()
			receiver := in[0]
			delegate := NewDelegate(receiver, method, in[1:])
			defer a.finallyProcessed(delegate, methodLocation)
			if !a.beforeProcessed(delegate, methodLocation) {
				return delegate.Result
			}
			delegate.Result = receiver.MethodByName(method.Name).Call(in[1:])
			a.afterProcessed(delegate, methodLocation)
			return delegate.Result

		}).Interface())
		a.hasProxy[methodLocation] = true
	}
}

func (a *aspect) RegisterProxy(proxy Proxy) error {
	a.Lock()
	defer a.Unlock()
	//a.units = append(a.units, proxy)
	if _, has := a.units[proxy.GetProxyName()]; has {
		return fmt.Errorf("proxy %s is exist", proxy.GetProxyName())
	}
	a.units[proxy.GetProxyName()] = proxy
	a.unitNames = append(a.unitNames, proxy.GetProxyName())
	return nil
}

func (a *aspect) beforeProcessed(delegate *Delegate, methodLocation string) bool {
	a.RLock()
	defer a.RUnlock()
	isDone := true
	a.unitsRun(true, methodLocation, func(subaspect Proxy, needExitRange *bool) {
		isDone = subaspect.Before(delegate, methodLocation)
		if !isDone {
			*needExitRange = true
		}
	})
	return isDone
}

func (a *aspect) afterProcessed(delegate *Delegate, methodLocation string) {
	a.RLock()
	defer a.RUnlock()
	a.unitsRun(false, methodLocation, func(proxy Proxy, needExitRange *bool) {
		proxy.After(delegate, methodLocation)
	})
}

func (a *aspect) finallyProcessed(delegate *Delegate, methodLocation string) {
	a.RLock()
	defer a.RUnlock()
	a.unitsRun(false, methodLocation, func(proxy Proxy, needExitRange *bool) {
		proxy.Finally(delegate, methodLocation)
	})
}

func (a *aspect) unitsRun(inOrder bool, methodLocation string, fn func(proxy Proxy, needExitRange *bool)) {
	var (
		i        int
		canRange func(int) bool
		nextIdx  func(*int)
	)
	if inOrder {
		i = 0
		canRange = func(j int) bool {
			return len(a.unitNames) > j
		}
		nextIdx = func(j *int) {
			*j++
		}
	} else {
		i = len(a.unitNames) - 1
		canRange = func(j int) bool {
			return j >= 0
		}
		nextIdx = func(j *int) {
			*j--
		}
	}

	needExitRange := false

	for ; canRange(i) && !needExitRange; nextIdx(&i) {
		unit := a.units[a.unitNames[i]]
		if !unit.IsMatch(methodLocation) {
			continue
		}
		fn(unit, &needExitRange)
	}

}
