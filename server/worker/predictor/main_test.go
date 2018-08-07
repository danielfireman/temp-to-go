package main

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/danielfireman/temp-to-go/server/tsmongo"

	"github.com/globalsign/mgo/dbtest"
)

var mongoDB dbtest.DBServer

func TestPredictor(t *testing.T) {
	tempDir, _ := ioutil.TempDir("", "predictor_testing")
	mongoDB.SetPath(tempDir)
	// Wipe already closes all sessions.
	defer func() { mongoDB.Wipe() }()

	fmt.Println("Before mgo session")
	_ = mongoDB.Session()
	fmt.Println("After")

	_, err := tsmongo.Dial("localhost:27017")
	if err != nil {
		t.Error(err)
	}
}
