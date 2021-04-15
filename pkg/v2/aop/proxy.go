package aop

import (
	"errors"
	"fmt"
	"strings"
)

type Proxy interface {
	GetProxyName() string
	Before(delegate *Delegate, methodLocation string) bool
	After(delegate *Delegate, methodLocation string)
	Finally(delegate *Delegate, methodLocation string)
	IsMatch(methodLocation string) bool
}

func NewProxy(
	proxyName string,
	beforeFn func(delegate *Delegate, methodLocation string) bool,
	afterFn func(delegate *Delegate, methodLocation string),
	finallyFn func(delegate *Delegate, methodLocation string),
	isMatch func(methodLocation string) bool,
) (Proxy, error) {
	if strings.EqualFold(strings.TrimSpace(proxyName), "") {
		return nil, errors.New("proxyName is null")
	}
	p := &proxy{
		name:      proxyName,
		beforeFn:  beforeFn,
		afterFn:   afterFn,
		finallyFn: finallyFn,
		isMatch:   isMatch,
	}
	return Proxy(p), nil
}

type proxy struct {
	name      string
	beforeFn  func(delegate *Delegate, methodLocation string) bool
	afterFn   func(delegate *Delegate, methodLocation string)
	finallyFn func(delegate *Delegate, methodLocation string)
	isMatch   func(methodLocation string) bool
}

func (p proxy) GetProxyName() string {
	return p.name
}

func (p proxy) Before(delegate *Delegate, methodLocation string) bool {
	if p.beforeFn == nil {
		return true
	}
	return p.beforeFn(delegate, methodLocation)
}

func (p proxy) After(delegate *Delegate, methodLocation string) {
	if p.afterFn != nil {
		p.afterFn(delegate, methodLocation)
	}
}

func (p proxy) Finally(delegate *Delegate, methodLocation string) {
	if p.finallyFn != nil {
		p.finallyFn(delegate, methodLocation)
	}
}

func (p proxy) IsMatch(methodLocation string) bool {
	if p.isMatch == nil {
		fmt.Printf("%s proxy can not match delegate\n", p.name)
		return false
	}
	return p.isMatch(methodLocation)
}
