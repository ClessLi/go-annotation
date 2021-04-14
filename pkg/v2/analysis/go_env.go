package analysis

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
	"sync"
)

var (
	singletonGoEnv goEnv
	onceForGoEnv   = new(sync.Once)
)

type goEnv map[string]string

func (g goEnv) get(k string) string {
	if v, has := g[k]; has {
		return v
	}
	return ""
}

func getSingletonGoEnv() goEnv {
	onceForGoEnv.Do(func() {
		cmd := exec.Command("go", "env")
		output, err := cmd.CombinedOutput()
		if err != nil {
			panic(err)
		}
		rd := bufio.NewReader(bytes.NewReader(output))
		var (
			isEnd = false
			line  = make([]byte, 0)
		)
		env := make(goEnv)
		for !isEnd {
			line, isEnd, err = rd.ReadLine()
			if err != nil {
				if strings.EqualFold(err.Error(), "EOF") {
					isEnd = true
					continue
				}
				panic(err)
			}
			fields := strings.Split(strings.TrimSpace(string(line)), " ")
			if len(fields) == 2 && strings.EqualFold(fields[0], "set") {
				kv := strings.Split(fields[1], "=")
				switch len(kv) {
				case 1:
					env[kv[0]] = ""
				case 2:
					env[kv[0]] = kv[1]
				}
			}
		}
		if len(env) < 1 {
			panic("get go env failed")
		}
		singletonGoEnv = env
	})
	return singletonGoEnv
}
