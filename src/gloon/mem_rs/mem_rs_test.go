package mem_rs

import (
	"testing"
)

func TestPutGetDel(t *testing.T) {
	r := Create()
	err := r.Clear()
	if err != nil {
		t.Error("Clear()", err)
	}
	err = r.PutVal(1, "foo.bar", "127.0.0.1")
	if err != nil {
		t.Error("r.PutVal()", err.Error())
	}
	vals, err := r.GetAll(1, "foo.bar")
	if err != nil {
		t.Error("r.GetVal()", err)
	}
	if vals[0] != "127.0.0.1" {
		t.Errorf("Got value %s -- expected 127.0.0.1", vals[0])
	}
	err = r.DelKey(1, "foo.bar")
	if err != nil {
		t.Error("r.Del()", err)
	}
	vals, err = r.GetAll(1, "foo.bar")
	if err != nil {
		t.Error("r.GetVal()", err)
	}
	if len(vals) != 0 {
		t.Errorf("Expected empty value. Got %s", vals[0])
	}
}

func TestMultiVals(t *testing.T) {
	r := Create()
	err := r.Clear()
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
	vals, err = r.GetAll(1, "foo.bar")
	if err != nil {
		t.Error("r.GetVal()", err)
	}
	if vals[0] != "127.0.0.2" {
		t.Errorf("Got %s, expected 127.0.0.2", vals[0])
	}
	err = r.DelVal(1, "foo.bar", "127.0.0.2")
	if err != nil {
		t.Error("r.DelValue()", err)
	}
	vals, err = r.GetAll(1, "foo.bar")
	if err != nil {
		t.Error("r.GetVal()", err)
	}
	if len(vals) != 0 {
		t.Errorf("Got %#v, expected empty value", vals)
	}
}
