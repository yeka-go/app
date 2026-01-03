package openapi_test

import (
	"testing"

	"github.com/yeka-go/app/cmd/goapp/internal/openapi"
)

func TestMap(t *testing.T) {
	type Data struct {
		Key   string
		Value string
	}

	m := openapi.Map[Data, string]{}
	d := Data{Key: "a", Value: "A"}
	m.Push(d, d.Key)
	d = Data{Key: "b", Value: "B"}
	m.Push(d, d.Key)

	m.Get("a").Value = "C"
	if m.Length() != 2 {
		t.Error("invalid length")
	}
	a := m.Get("a")
	if a.Key != "a" || a.Value != "C" {
		t.Error("invalid get a")
	}
	b := m.Get("b")
	if b.Key != "b" || b.Value != "B" {
		t.Error("invalid get b")
	}
}
