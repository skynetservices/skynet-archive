package bsonrpc

import (
	"bytes"
	"labix.org/v2/mgo/bson"
	"net/rpc"
	"testing"
)

func TestEncode(t *testing.T) {
}

func TestDecode(t *testing.T) {
	req := rpc.Request{
		ServiceMethod: "Foo.Bar",
		Seq:           3,
	}

	b, err := bson.Marshal(req)

	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(b)
	dec := NewDecoder(buf)

	r := new(rpc.Request)
	err = dec.Decode(r)

	if err != nil {
		t.Fatal(err)
	}

	if *r != req {
		t.Fatal("Values don't match")
	}
}

func TestDecodeReadsOnlyOne(t *testing.T) {
	req := rpc.Request{
		ServiceMethod: "Foo.Bar",
		Seq:           3,
	}

	type T struct {
		Value string
	}

	tv := T{"test"}

	b, err := bson.Marshal(req)

	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(b)

	b, err = bson.Marshal(tv)

	if err != nil {
		t.Fatal(err)
	}

	buf.Write(b)
	dec := NewDecoder(buf)

	r := new(rpc.Request)
	err = dec.Decode(r)

	if *r != req {
		t.Fatal("Values don't match")
	}

	if err != nil {
		t.Fatal(err)
	}

	// We should be able to read a second message off this io.Reader
	tmp := new(T)
	err = dec.Decode(tmp)

	if err != nil {
		t.Fatal(err)
	}

	if *tmp != tv {
		t.Fatal("Values don't match")
	}

}
