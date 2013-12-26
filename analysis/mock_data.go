package main

import (
	"github.com/ziutek/mymysql/mysql"
	"log"
	"net"
	"time"
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

func (self *mockConnection) Clone() mysql.Conn {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil
}

func (self *mockConnection) SetTimeout(time.Duration) {
	log.Println("Called a method that is not implemented in the mock database object.")
}

func (self *mockConnection) Connect() error {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil
}

func (self *mockConnection) NetConn() net.Conn {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil
}

func (self *mockConnection) SetDialer(mysql.Dialer) {
	log.Println("Called a method that is not implemented in the mock database object.")
	return
}

func (self *mockConnection) Close() error {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil
}

func (self *mockConnection) IsConnected() bool {
	log.Println("Called a method that is not implemented in the mock database object.")
	return false
}

func (self *mockConnection) Reconnect() error {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil
}

func (self *mockConnection) Use(dbname string) error {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil
}

func (self *mockConnection) Register(sql string) {
	log.Println("Called a method that is not implemented in the mock database object.")
}

func (self *mockConnection) SetMaxPktSize(new_size int) int {
	log.Println("Called a method that is not implemented in the mock database object.")
	return 0
}

func (self *mockConnection) NarrowTypeSet(narrow bool) {
	log.Println("Called a method that is not implemented in the mock database object.")
}

func (self *mockConnection) FullFieldInfo(full bool) {
	log.Println("Called a method that is not implemented in the mock database object.")
}

func (self *mockConnection) Begin() (mysql.Transaction, error) {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil, nil
}
