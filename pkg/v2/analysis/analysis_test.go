package analysis

import (
	"go/parser"
	"go/token"
	"io/ioutil"
	"testing"
)

type TestTransact struct {
}

//@Transactional
func (*TestTransact) Before() {
}

func TestAnalyser_ScanFuncDecl(t *testing.T) {
	fileName := `analysis_test.go`
	bt, _ := ioutil.ReadFile(fileName)
	src := string(bt)
	fSet := token.NewFileSet()
	f, err := parser.ParseFile(fSet, fileName, src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	analyser := NewAnalyser()
	info := analyser.ScanFuncDecl(f, "Transactional")
	for _, recvName := range info.GetRecvNames() {
		for _, methodInfo := range info.GetRecv(recvName).Methods {
			if methodInfo.HasAnnotation["Transactional"] {
				t.Logf("%v.%v has 'Transactional' annotation.", recvName, methodInfo.MethodName)
			} else {
				t.Logf("%v.%v has not 'Transactional' annotation.", recvName, methodInfo.MethodName)
			}
		}
	}
}
