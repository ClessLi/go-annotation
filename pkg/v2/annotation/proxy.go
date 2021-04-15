package annotation

type AnnotatedMethodProxy interface {
	GetProxyName() string
	Before(delegate AnnotatedMethod) bool
	After(delegate AnnotatedMethod)
	Finally(delegate AnnotatedMethod)
}
