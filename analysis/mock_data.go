package main

import (
	"github.com/ziutek/mymysql/mysql"
	"log"
)

type mockConnection struct {
}

func (self *mockConnection) Start(sql string, params ...interface{}) (mysql.Result, error) {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil, nil
}

func (self *mockConnection) Prepare(sql string) (mysql.Stmt, error) {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil, nil
}

func (self *mockConnection) Ping() error {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil
}

func (self *mockConnection) ThreadId() uint32 {
	log.Println("Called a method that is not implemented in the mock database object.")
	return 0
}

func (self *mockConnection) Escape(txt string) string {
	log.Println("Called a method that is not implemented in the mock database object.")
	return ""
}

func (self *mockConnection) Query(sql string, params ...interface{}) ([]mysql.Row, mysql.Result, error) {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil, nil, nil
}

func (self *mockConnection) QueryFirst(sql string, params ...interface{}) (mysql.Row, mysql.Result, error) {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil, nil, nil
}

func (self *mockConnection) QueryLast(sql string, params ...interface{}) (mysql.Row, mysql.Result, error) {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil, nil, nil
}
