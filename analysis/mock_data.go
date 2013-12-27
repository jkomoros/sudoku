package main

import (
	"encoding/csv"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type mockConnection struct {
}

type mockResult struct {
	reader        *csv.Reader
	isSolvesTable bool
}

func (self *mockConnection) Start(sql string, params ...interface{}) (mysql.Result, error) {

	isSolvesTable := false
	filename := "mock_data/puzzles_data.csv"

	sql = fmt.Sprintf(sql, params...)

	if strings.Contains(sql, config.SolvesTable) {
		isSolvesTable = true
		filename = "mock_data/solves_data.csv"
	}

	file, err := os.Open(filename)

	if err != nil {
		log.Fatal("Couldn't open the file of mock data: ", filename)
		os.Exit(1)
	}

	//We'd normally call defer file.Close() here, but we can't because we still have to vend the rows.

	return &mockResult{csv.NewReader(file), isSolvesTable}, nil
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

	data, _ := self.reader.Read()

	if data == nil {
		return nil, nil
	}

	if self.isSolvesTable {
		if len(data) != 4 {
			log.Fatal("The data in the mock solves table should have four items but at least one row doesn't")
			os.Exit(1)
		}

		id, _ := strconv.Atoi(data[1])
		solveTime, _ := strconv.Atoi(data[2])
		penaltyTime, _ := strconv.Atoi(data[3])

		return mysql.Row{data[0], id, solveTime, penaltyTime}, nil
	} else {
		if len(data) != 4 {
			log.Fatal("The data in the mock puzzles table should have four items but at least one row doesn't.")
			os.Exit(1)
		}
		id, _ := strconv.Atoi(data[0])
		difficulty, _ := strconv.Atoi(data[1])

		return mysql.Row{id, difficulty, data[2], data[3]}, nil
	}

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
