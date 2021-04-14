package analysis

import (
	"fmt"
	"go/ast"
	"strings"
)

type Analyzer interface {
	ScanFuncDecl(file *ast.File, targetAnnotation string) PackageInfo
	ScanMethodByClass(object interface{}, targetAnnotation string) PackageInfo
}

type analyzer struct {
	commentPrefix    string
	annotationPrefix string
}

func NewAnalyzer() Analyzer {
	return &analyzer{
		commentPrefix:    commentsPrefix,
		annotationPrefix: annotationPrefix,
	}
}

func (a analyzer) ScanFuncDecl(file *ast.File, targetAnnotation string) PackageInfo {
	info := newPkgInfo(file.Name.String())
	for _, d := range file.Decls {
		switch decl := d.(type) {
		case *ast.FuncDecl:
			if decl.Doc != nil && a.isContainAnnotation(decl.Doc.List, targetAnnotation) {
				a.analysisAnnotation(decl, info, targetAnnotation)
			}
		}
	}
	return PackageInfo(info)
}

func (a analyzer) ScanMethodByClass(object interface{}, targetAnnotation string) PackageInfo {
	f, err := GetSingletonObjectAnalyzerInstance().AnalysisObjectToAstFiles(object)
	if err != nil || f == nil || len(f) < 1 {
		return nil
	}

	info := newPkgInfo(f[0].Name.String())
	for _, file := range f {
		for _, d := range file.Decls {
			doCaseFuncDeclPtr(d, func(fd *ast.FuncDecl) bool {
				if fd.Doc != nil && a.isContainAnnotation(fd.Doc.List, targetAnnotation) {
					a.analysisAnnotation(fd, info, targetAnnotation)
					return true
				}
				return false
			})
		}
	}
	return PackageInfo(info)
}

func (a analyzer) isContainAnnotation(lines []*ast.Comment, targetAnnotation string) bool {
	for _, l := range lines {
		c := strings.TrimSpace(strings.TrimLeft(l.Text, a.commentPrefix))
		annotation := strings.TrimLeft(c, a.annotationPrefix)
		if annotation == targetAnnotation {
			return true
		}
	}
	return false
}

func (a analyzer) analysisAnnotation(decl *ast.FuncDecl, info *pkgInfo, annotation string) {
	if info.Receivers == nil {
		info.Receivers = make(map[string]*RecvInfo)
	}

	if info.Funcs == nil {
		info.Funcs = make(map[string]*FuncInfo)
	}

	if decl.Recv != nil {
		field := decl.Recv.List[0]
		methodName := decl.Name.String()
		a.analysisAnnotationToMethod(field, info, methodName, annotation)
	} else {
		funcName := decl.Name.String()
		a.analysisAnnotationToFunc(info, funcName, annotation)

	}
}

func (a analyzer) analysisAnnotationToMethod(field *ast.Field, info *pkgInfo, methodName string, annotation string) {
	switch f := field.Type.(type) {
	case *ast.StarExpr:
		recvName := fmt.Sprintf("%v", f.X)

		if info.Receivers[recvName] == nil {
			info.Receivers[recvName] = newRecvInfo(info.GetPackageName(), recvName)
		}

		info.Receivers[recvName].SetMethod(methodName, annotation)
	}

}

func (a analyzer) analysisAnnotationToFunc(info *pkgInfo, funcName string, annotation string) {
	if _, has := info.Funcs[funcName]; !has {
		info.Funcs[funcName] = newFuncInfo(info.GetPackageName(), funcName)
	}
	info.Funcs[funcName].SetAnnotation(annotation)
}

//func (a analyzer) scanFileByClass(object interface{}) *ast.File {
//	gomodFilePath := os.Getenv("GOMOD")
//	gosumFilePath := filepath.Join(filepath.Dir(gomodFilePath), "go.sum")
//	pkgPath := reflect.TypeOf(object).PkgPath()
//
//	gomodPkgPath := os.Getenv("GOMODCACHE")
//	gopath := os.Getenv("GOPATH")
//}

type PackageInfo interface {
	GetPackageName() string
	GetRecvNames() []string
	GetRecv(recvName string) *RecvInfo
	GetFuncs() map[string]*FuncInfo
}

type pkgInfo struct {
	PkgName string
	//RecvMethods map[string][]MethodInfo // key RecvName
	Receivers map[string]*RecvInfo
	Funcs     map[string]*FuncInfo
}

func (p pkgInfo) GetPackageName() string {
	return p.PkgName
}

func (p pkgInfo) GetRecvNames() []string {
	names := make([]string, 0)
	for s := range p.Receivers {
		names = append(names, s)
	}
	return names
}

func (p pkgInfo) GetRecv(recvName string) *RecvInfo {
	if p.Receivers == nil {
		panic("pkgInfo.Receivers is nil ptr.")
	}
	if _, has := p.Receivers[recvName]; has {
		return p.Receivers[recvName]
	}
	return nil
}

func (p pkgInfo) GetFuncs() map[string]*FuncInfo {
	return p.Funcs
}

func newPkgInfo(pkgName string) *pkgInfo {
	return &pkgInfo{
		PkgName:   pkgName,
		Receivers: make(map[string]*RecvInfo),
		Funcs:     make(map[string]*FuncInfo),
	}
}

type RecvInfo struct {
	PkgName  string
	RecvName string
	Methods  map[string]*MethodInfo
}

func newRecvInfo(pkgName, recvName string) *RecvInfo {
	return &RecvInfo{
		PkgName:  pkgName,
		RecvName: recvName,
		Methods:  make(map[string]*MethodInfo),
	}
}

func (r *RecvInfo) SetMethod(methodName string, annotations ...string) {
	if _, has := r.Methods[methodName]; !has {
		r.Methods[methodName] = newMethodInfo(r.PkgName, r.RecvName, methodName)
	}

	for _, annotation := range annotations {
		r.Methods[methodName].SetAnnotation(annotation)
	}
}

type MethodInfo struct {
	PkgName       string
	RecvName      string
	MethodName    string
	HasAnnotation map[string]bool
}

func newMethodInfo(pkgName, recvName, methodName string) *MethodInfo {
	return &MethodInfo{
		PkgName:       pkgName,
		RecvName:      recvName,
		MethodName:    methodName,
		HasAnnotation: make(map[string]bool),
	}
}

func (m *MethodInfo) SetAnnotation(annotation string) {
	m.HasAnnotation[annotation] = true
}

type FuncInfo struct {
	PkgName       string
	FuncName      string
	HasAnnotation map[string]bool
}

func newFuncInfo(pkgName, funcName string) *FuncInfo {
	return &FuncInfo{
		PkgName:       pkgName,
		FuncName:      funcName,
		HasAnnotation: make(map[string]bool),
	}
}

func (f *FuncInfo) SetAnnotation(annotation string) {
	f.HasAnnotation[annotation] = true
}
