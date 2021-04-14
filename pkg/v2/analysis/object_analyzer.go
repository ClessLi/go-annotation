package analysis

import (
	"bufio"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

type ObjectAnalyzer interface {
	AnalysisObjectToAstFiles(object interface{}) ([]*ast.File, error)
}

var (
	SingletonObjectAnalyzer ObjectAnalyzer
	onceForObjectAnalyzer   = new(sync.Once)
)

func GetSingletonObjectAnalyzerInstance() ObjectAnalyzer {
	onceForObjectAnalyzer.Do(func() {
		var err error
		goModPath := getSingletonGoEnv().get("GOMOD")
		//if !filepath.IsAbs(goModPath) {
		//	goModPath, err = filepath.Abs(goModPath)
		//	if err != nil {
		//		panic(fmt.Sprintf("abs go.mod filepath error, case by: %s", err))
		//	}
		//}
		srcPath := filepath.Dir(goModPath)
		goModModule, goModPkgSrcPaths, err := analysisGoMod(goModPath)
		if err != nil {
			panic(fmt.Sprintf("analysis go.mod error, case by: %s", err))
		}
		goModPkgSrcPaths[goModModule] = srcPath
		SingletonObjectAnalyzer = &objectAnalyzer{
			goModPkgSrcPaths: goModPkgSrcPaths,
		}
	})
	return SingletonObjectAnalyzer
}

func analysisGoMod(goModPath string) (moduleName string, pkgPaths map[string]string, err error) {
	moduleName, err = parseGoModModuleName(goModPath)
	if err != nil {
		return "", nil, err
	}
	goSumPath := filepath.Join(filepath.Dir(goModPath), "go.sum")
	pkgPaths, err = parseGoModPkgPaths(goSumPath)
	if err != nil {
		return "", nil, err
	}
	return moduleName, pkgPaths, nil
}

func parseGoModPkgPaths(goSumPath string) (map[string]string, error) {
	rd, closeFn, err := readFile(goSumPath)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	//goModCachePath := os.Getenv("GOMODCACHE")
	goModCachePath := getSingletonGoEnv().get("GOMODCACHE")
	if strings.TrimSpace(goModCachePath) == "" {
		return nil, errors.New("$GOPATH is null")
	}
	isEnd := false
	line := make([]byte, 0)
	pkgPaths := make(map[string]string)

	for !isEnd {
		line, isEnd, err = rd.ReadLine()
		if err != nil {
			if strings.EqualFold(err.Error(), "EOF") {
				isEnd = true
				continue
			}
			return nil, err
		}
		if len(line) == 0 {
			continue
		}

		fields := strings.Split(strings.TrimSpace(string(line)), " ")
		if len(fields) != 3 || strings.HasSuffix(fields[1], "/go.mod") {
			continue
		}

		pkgName := fields[0]
		pkgVersion := fields[1]
		pkgSrcPath, err := parsePkgSrcPath(goModCachePath, pkgName, pkgVersion)
		if err != nil {
			//return nil, err
			continue
		}
		pkgPaths[pkgName] = pkgSrcPath
	}
	return pkgPaths, nil
}

func parsePkgSrcPath(goModCachePath, pkgName, pkgVersion string) (string, error) {
	pkgPath := formatUpper(pkgName)
	pkgPath = filepath.Join(strings.Split(pkgPath, "/")...)
	pkgSrcPath := filepath.Join(goModCachePath, pkgPath+"@"+pkgVersion)
	stat, err := os.Stat(pkgSrcPath)
	if err != nil {
		return "", err
	}
	if stat.IsDir() {
		return pkgSrcPath, nil
	}
	return "", fmt.Errorf("parse pkg path %s failed, it's not in local go mod libs", pkgName)
}

func formatUpper(o string) string {
	bs := make([]byte, 0)
	copy(bs, o)
	for i, s := range bs {
		if 64 < s && s < 91 {
			bs = append(bs[:i], append([]byte{"!"[0], s - 32}, bs[i+1:]...)...)
		}
	}
	return string(bs)
}

func readFile(filepath string) (*bufio.Reader, func(), error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, func() {}, err
	}

	rd := bufio.NewReader(f)
	return rd, func() {
		err := f.Close()
		if err != nil {
			fmt.Printf("close file %s error: %s", filepath, err)
		}
	}, nil
}

func parseGoModModuleName(goModPath string) (string, error) {
	rd, closeFn, err := readFile(goModPath)
	if err != nil {
		return "", err
	}
	defer closeFn()

	isEnd := false
	line := make([]byte, 0)

	for !isEnd {
		line, isEnd, err = rd.ReadLine()
		if err != nil {
			if strings.EqualFold(err.Error(), "EOF") {
				isEnd = true
				continue
			}
			return "", err
		}

		if len(line) == 0 {
			continue
		}
		fields := strings.Split(strings.TrimSpace(string(line)), " ")
		if len(fields) < 2 || !strings.EqualFold(fields[0], "module") {
			continue
		}
		return fields[len(fields)-1], nil
	}
	err = errors.New("can not match module name")
	return "", err
}

type objectAnalyzer struct {
	goModPkgSrcPaths map[string]string
}

func (o objectAnalyzer) AnalysisObjectToAstFiles(object interface{}) ([]*ast.File, error) {
	refObject := reflect.TypeOf(object)
	if refObject.Kind() == reflect.Ptr {
		refObject = refObject.Elem()
	}
	pkgPath := refObject.PkgPath()
	pkgFullPath := o.analysisPkgPath(pkgPath)
	return o.analysisToAstFiles(pkgFullPath, refObject.Name())
}

func (o objectAnalyzer) analysisPkgPath(pkgPath string) string {
	dirSlice := strings.Split(pkgPath, "/")
	var pkgSrcPath, pkgSubPath string
	for i := 0; i < len(dirSlice); i++ {
		pkgSrcPath = strings.Join(dirSlice[:i+1], "/")
		pkgSubPath = filepath.Join(dirSlice[i+1:]...)
		if _, has := o.goModPkgSrcPaths[pkgSrcPath]; has {
			pkgSrcFullPath := o.goModPkgSrcPaths[pkgSrcPath]
			pkgFullPath := filepath.Join(pkgSrcFullPath, pkgSubPath)
			return pkgFullPath
		}
	}
	return ""
}

func (o objectAnalyzer) analysisToAstFiles(pkgFullPath, objectName string) ([]*ast.File, error) {
	if strings.TrimSpace(pkgFullPath) == "" {
		return nil, fmt.Errorf("can not match the pkg file for object %s", objectName)
	}
	filePaths, err := filepath.Glob(filepath.Join(pkgFullPath, "*"))
	if err != nil {
		return nil, err
	}

	fs := make([]*ast.File, 0)
	for _, filePath := range filePaths {
		s, err := os.Stat(filePath)
		if err != nil {
			return nil, err
		}
		if s.IsDir() || !strings.EqualFold(path.Ext(filePath), goFileSuffix) {
			continue
		}
		//bt, _ := ioutil.ReadFile(filePath)
		//src := string(bt)
		fSet := token.NewFileSet()
		f, err := parser.ParseFile(fSet, filePath, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		if o.isContainAstFile(f, objectName) {
			fs = append(fs, f)
		}

	}
	return fs, nil
}

func (o objectAnalyzer) isContainAstFile(f *ast.File, objectName string) bool {
	do := func(fd *ast.FuncDecl) bool {
		if fd.Recv != nil {
			field := fd.Recv.List[0]
			if ft, isType := field.Type.(*ast.StarExpr); isType {
				return strings.EqualFold(objectName, fmt.Sprintf("%v", ft.X))
			}
		}
		return false
	}

	for _, d := range f.Decls {
		if doCaseFuncDeclPtr(d, do) {
			return true
		}
	}
	return false

}

func doCaseFuncDeclPtr(d ast.Decl, do func(fd *ast.FuncDecl) bool) bool {
	switch decl := d.(type) {
	case *ast.FuncDecl:
		return do(decl)
	}
	return false
}
