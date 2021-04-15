package annotation

import (
	"github.com/ClessLi/go-annotation/pkg/v2/analysis"
	"github.com/ClessLi/go-annotation/pkg/v2/aop"
	"reflect"
)

type Annotation interface {
	RegisterAnnotatedObject(object interface{})
	RegisterAnnotatedObjectProxy(proxy AnnotatedMethodProxy) error
	GetAnnotatedMethodInfos() map[string]*analysis.MethodInfo
}

type annotation struct {
	targetAnnotation string
	aopAspect        aop.ProxiesAspect
	annotatedMethods map[string]*analysis.MethodInfo
	analyzer         analysis.Analyzer
}

func newAnnotation(targetAnnotation string, aspect aop.ProxiesAspect, analyzer analysis.Analyzer) Annotation {
	a := &annotation{
		targetAnnotation: targetAnnotation,
		aopAspect:        aspect,
		annotatedMethods: make(map[string]*analysis.MethodInfo),
		analyzer:         analyzer,
	}
	return Annotation(a)
}

func NewAnnotation(targetAnnotation string) Annotation {
	return newAnnotation(targetAnnotation, aop.GetSingletonAspectInstance(), analysis.NewAnalyzer())
}

func (a *annotation) RegisterAnnotatedObject(object interface{}) {
	objectType := reflect.TypeOf(object)
	if objectType.Kind() == reflect.Ptr {
		objectType = objectType.Elem()
	}
	//pkgPath := objectType.PkgPath()
	pkgInfo := a.analyzer.ScanMethodByClass(object, a.targetAnnotation)
	classInfo := pkgInfo.GetRecv(objectType.Name())
	if classInfo != nil {
		a.aopAspect.RegisterDelegate(objectType)
		for _, methodInfo := range classInfo.Methods {
			methodLocation := methodInfo.PkgName + "." + methodInfo.RecvName + "." + methodInfo.MethodName
			a.annotatedMethods[methodLocation] = methodInfo
		}
	}
}

func (a *annotation) RegisterAnnotatedObjectProxy(proxy AnnotatedMethodProxy) error {
	p, err := aop.NewProxy(
		proxy.GetProxyName(),
		func(delegate *aop.Delegate, methodLocation string) bool {
			am := newAnnotatedMethod(delegate, methodLocation)
			return proxy.Before(am)
		},
		func(delegate *aop.Delegate, methodLocation string) {
			am := newAnnotatedMethod(delegate, methodLocation)
			proxy.After(am)
		},
		func(delegate *aop.Delegate, methodLocation string) {
			am := newAnnotatedMethod(delegate, methodLocation)
			proxy.Finally(am)
		},
		func(methodLocation string) bool {
			if methodInfo, has := a.annotatedMethods[methodLocation]; has && methodInfo != nil {
				return true
			}
			return false
		},
	)
	if err != nil {
		return err
	}
	return a.aopAspect.RegisterProxy(p)
}

func (a annotation) GetAnnotatedMethodInfos() map[string]*analysis.MethodInfo {
	return a.annotatedMethods
}
