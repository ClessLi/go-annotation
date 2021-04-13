package main

import (
	"github.com/ClessLi/go-annotation/example"
	"github.com/ClessLi/go-annotation/pkg/annotation/transaction"
	"github.com/go-xorm/xorm"
)

// go build -gcflags=-l main.go
// main
func main() {
	scanPath := `F:\GOPATH\src\github.com\handsomestWei\go-annotation\example`
	transaction.NewTransactionManager(transaction.TransactionConfig{ScanPath: scanPath}).RegisterDao(new(example.ExampleDao))

	dao := new(example.ExampleDao)
	dao.Select()
	dao.Update(new(xorm.Session), "") // auto commit
	dao.Delete(new(xorm.Session))     // handle fail and auto rollback
}
