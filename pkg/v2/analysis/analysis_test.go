package analysis

import (
	"fmt"
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

type TestTransact struct {
}

//@Transactional
func (*TestTransact) Before() {
}

//@EventBus
//@Transactional
func (*TestTransact) After() {
}

func (t *TestTransact) Finally() {
}

func TestAnalyser_ScanFuncDecl(t *testing.T) {
	fileName := `analysis_test.go`
	//bt, _ := ioutil.ReadFile(fileName)
	//src := string(bt)
	fSet := token.NewFileSet()
	f, err := parser.ParseFile(fSet, fileName, nil, parser.ParseComments)
	//f, err := parser.ParseFile(fSet, fileName, src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	analyzer := NewAnalyzer()
	info := analyzer.ScanFuncDecl(f, "Transactional")
	for _, recvName := range info.GetRecvNames() {
		for _, methodInfo := range info.GetRecv(recvName).Methods {
			if methodInfo.HasAnnotation("Transactional") {
				t.Logf("%v.%v has 'Transactional' annotation.", recvName, methodInfo.MethodName)
				t.Logf("%v", methodInfo)
			} else {
				t.Logf("%v.%v has not 'Transactional' annotation.", recvName, methodInfo.MethodName)
			}
		}
	}
}

func TestAnalyzer_ScanMethodByClass(t *testing.T) {
	test := new(TestTransact)
	analyzer := NewAnalyzer()
	info := analyzer.ScanMethodByClass(test, "Transactional")
	for _, recvName := range info.GetRecvNames() {
		for _, methodInfo := range info.GetRecv(recvName).Methods {
			if methodInfo.HasAnnotation("Transactional") {
				t.Logf("%v.%v has 'Transactional' annotation.", recvName, methodInfo.MethodName)
				t.Logf("%v", methodInfo)
			} else {
				t.Logf("%v.%v has not 'Transactional' annotation.", recvName, methodInfo.MethodName)
			}
		}
	}
}

func TestGetPkgPath(t *testing.T) {
	//a := Analyzer(analyzer{
	//	commentPrefix:    commentsPrefix,
	//	annotationPrefix: annotationPrefix,
	//})
	a := NewAnalyzer()
	fmt.Printf("%s\n", reflect.TypeOf(a).Elem().PkgPath())
	fmt.Printf("%s\n", reflect.TypeOf(a).Elem().Name())
}
