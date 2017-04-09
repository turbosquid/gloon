package redis_rs

import (
	"testing"
)

func TestPutGetDel(t *testing.T) {
	r, err := Create("localhost:6379,2,test")
	if err != nil {
		t.Error("CreateRecordStore", err.Error())
	}
	err = r.Clear()
	if err != nil {
		t.Error("Clear()", err)
	}
	err = r.PutVal(1, "foo.bar", "127.0.0.1")
	if err != nil {
		t.Error("r.PutVal()", err.Error())
	}
	val, err := r.GetVal(1, "foo.bar")
	if err != nil {
		t.Error("r.GetVal()", err)
	}
	if val != "127.0.0.1" {
		t.Errorf("Got value %s -- expected 127.0.0.1", val)
	}
	err = r.DelKey(1, "foo.bar")
	if err != nil {
		t.Error("r.Del()", err)
	}
	val, err = r.GetVal(1, "foo.bar")
	if err != nil {
		t.Error("r.GetVal()", err)
	}
	if val != "" {
		t.Errorf("Expected empty value", val)
	}
}

func TestMultiVals(t *testing.T) {
	r, err := Create("localhost:6379,2,test")
	if err != nil {
		t.Error("CreateRecordStore", err.Error())
	}
	err = r.Clear()
	if err != nil {
		t.Error("Clear()", err)
	}
	err = r.PutVal(1, "foo.bar", "127.0.0.1")
	if err != nil {
		t.Error("r.PutVal()", err.Error())
	}
	err = r.PutVal(1, "foo.bar", "127.0.0.2")
	if err != nil {
		t.Error("r.PutVal() second val", err.Error())
	}
	val, err := r.GetVal(1, "foo.bar")
	if err != nil {
		t.Error("r.GetVal()", err)
	}
	if val == "" {
		t.Error("r.GetVal()", "Unexpected empty value")
	}
	t.Logf("Got random entry: %s", val)
	vals, err := r.GetAll(1, "foo.bar")
	if err != nil {
		t.Error("r.GetAll()", err)
	}
	t.Logf("Got multiple Values: %#v", vals)
	if len(vals) != 2 {
		t.Errorf("Got %d values (%#v) -- expected 2", len(vals), vals)
	}
	err = r.DelVal(1, "foo.bar", "127.0.0.1")
	if err != nil {
		t.Error("r.DelValue()", err)
	}
	val, err = r.GetVal(1, "foo.bar")
	if err != nil {
		t.Error("r.GetVal()", err)
	}
	if val != "127.0.0.2" {
		t.Errorf("Got %s, expected 127.0.0.2", val)
	}
}
