package main

import (
	"github.com/ziutek/mymysql/mysql"
	"log"
	"net"
	"time"
)

type mockConnection struct {
}

type mockResult struct {
}

func (self *mockConnection) Start(sql string, params ...interface{}) (mysql.Result, error) {
	return &mockResult{}, nil
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
	//Just pretend everything worked correctly.
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

//Begin mockResult methods

func (self *mockResult) StatusOnly() bool {
	log.Println("Called a method that is not implemented in the mock database object.")
	return false
}

func (self *mockResult) ScanRow(mysql.Row) error {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil
}

func (self *mockResult) GetRow() (mysql.Row, error) {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil, nil
}

func (self *mockResult) MoreResults() bool {
	log.Println("Called a method that is not implemented in the mock database object.")
	return false
}

func (self *mockResult) NextResult() (mysql.Result, error) {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil, nil
}

func (self *mockResult) Fields() []*mysql.Field {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil
}

func (self *mockResult) Map(string) int {
	log.Println("Called a method that is not implemented in the mock database object.")
	return 0
}

func (self *mockResult) Message() string {
	log.Println("Called a method that is not implemented in the mock database object.")
	return ""
}

func (self *mockResult) AffectedRows() uint64 {
	log.Println("Called a method that is not implemented in the mock database object.")
	return 0
}

func (self *mockResult) InsertId() uint64 {
	log.Println("Called a method that is not implemented in the mock database object.")
	return 0
}

func (self *mockResult) WarnCount() int {
	log.Println("Called a method that is not implemented in the mock database object.")
	return 0
}

func (self *mockResult) MakeRow() mysql.Row {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil
}

func (self *mockResult) GetRows() ([]mysql.Row, error) {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil, nil
}

func (self *mockResult) End() error {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil
}

func (self *mockResult) GetFirstRow() (mysql.Row, error) {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil, nil
}

func (self *mockResult) GetLastRow() (mysql.Row, error) {
	log.Println("Called a method that is not implemented in the mock database object.")
	return nil, nil
}
