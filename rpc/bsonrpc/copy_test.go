package bsonrpc

import (
	"fmt"
	"launchpad.net/mgo/v2/bson"
	"testing"
)

func TestCopyStruct(t *testing.T) {
	var src = bson.M{
		"Hi":   "there",
		"What": []interface{}{"is", "up"},
	}
	type Dtyp struct {
		Hi   string
		What []string
	}
	var dst Dtyp
	CopyTo(src, &dst)
	if dst.Hi != "there" {
		t.Errorf("Expected %q, got %q", "there", dst.Hi)
	}
	fmt.Printf("%v\n", dst)
	t.Error("fine")
}
