package main

import (
	"fmt"
	"github.com/wqx/sparrow"
	// "github.com/wqx/sparrow/inject"
	// "github.com/go-martini/martini"
)

type DBConnection struct {
	Name string
}

func NewDBConnection(name string) *DBConnection {
	return &DBConnection{Name: name}
}

var DBConnectionMap map[string]*DBConnection

// Controller
func helloHandle1() {
	fmt.Println("helloHandle1")
	// return "Hello 1"
}

func helloHandle2() string {
	fmt.Println("helloHandle2")
	return "Hello 2"
}

func helloHandle3(dbMap map[string]*DBConnection) string {
	fmt.Printf("----DB: %s\n", dbMap["mpr_dbuser"].Name)
	return "Hello 3"
}

func main() {

	// Initial DB
	DBConnectionMap = make(map[string]*DBConnection)
	DBConnectionMap["mpr_dbuser"] = NewDBConnection("dbuser")

	sparrow := sparrow.Default()
	sparrow.Map(DBConnectionMap)

	sparrow.GET("/", func() string { return "Hello world" })
	sparrow.GET("/hello", helloHandle3, helloHandle1, helloHandle2)
	sparrow.ListenAndServe(":9090")

	//m := martini.Classic()
	//m.Get("/hello", helloHandle2, helloHandle1, helloHandle2)
	//m.RunOnAddr(":9090")
}
