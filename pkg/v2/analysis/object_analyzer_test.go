package analysis

import (
	"testing"
)

func TestGetSingletonObjectAnalyzerInstance(t *testing.T) {
	oa := GetSingletonObjectAnalyzerInstance()
	test := new(TestTransact)
	fs, err := oa.AnalysisObjectToAstFiles(test)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fs {
		t.Logf("%v", f)
	}
}
