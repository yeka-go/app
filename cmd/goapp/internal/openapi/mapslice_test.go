package openapi_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/yeka-go/app/cmd/goapp/internal/openapi"
)

var source = `{root: {satu: {nama: John, age: 17}, dua: [{nama: satu}, {nama: dua}]}}`
var additions = `{add1: {new: hello}, add2: [{name: john}, {name: jane}], add3: doe}`

func getAdditions(t *testing.T) (a, b, c any) {
	adds, err := openapi.LoadFromBytes([]byte(additions))
	if err != nil {
		t.Error("fail to load string:", err)
	}

	obj, ok1 := adds.GetPath("/add1")
	arr, ok2 := adds.GetPath("/add2")
	str, ok3 := adds.GetPath("/add3")

	if !ok1 || !ok2 || !ok3 {
		t.Error("unable to get str from object")
	}
	return obj, arr, str
}

func TestMapSlice_GetPath(t *testing.T) {
	ms := openapi.MapSlice{}
	err := ms.FromBytes([]byte(source))
	if err != nil {
		t.Error("unable to parse source")
	}

	// Test Path exists
	testData := []struct {
		Path           string
		ExpectedFound  bool
		ExpectedResult string
	}{
		{"/", true, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}]}}`},
		{"/root", true, `{"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}]}`},
		{"/root/satu", true, `{"nama": "John", "age": 17}`},
		{"/root/satu/nama", true, `"John"`},
		{"/root/dua", true, `[{"nama": "satu"}, {"nama": "dua"}]`},
		{"/root/dua/0", true, `{"nama": "satu"}`},
		{"/root/dua/0/nama", true, `"satu"`},
		{"/dua", false, `null`},
		{"/root/dua/nama", false, `null`},
	}
	for i, v := range testData {
		ok := ms.PathExists(v.Path)
		if ok != v.ExpectedFound {
			t.Errorf("expect \"%v\", got \"%v\" on test data #%v (path: %v)", v.ExpectedFound, ok, i+1, v.Path)
		}

		res, ok := ms.GetPath(v.Path)
		b, _ := yaml.MarshalWithOptions(res, yaml.JSON())
		str := strings.TrimSpace(string(b))
		if str != v.ExpectedResult {
			t.Errorf("expect `%v`, got `%v` on test data #%v (path: %v)", v.ExpectedResult, str, i+1, v.Path)
		}

	}
}

func TestMapSlice_AddPath(t *testing.T) {
	obj, arr, str := getAdditions(t)

	// Test Add to Path
	testData := []struct {
		Path           string
		Addition       any
		ExpectedError  bool
		ExpectedResult string
	}{
		{"/", obj, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}]}, "new": "hello"}`},
		{"/empat", obj, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}]}, "empat": {"new": "hello"}}`},
		{"/root", obj, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}], "new": "hello"}}`},
		{"/root/satu", obj, false, `{"root": {"satu": {"nama": "John", "age": 17, "new": "hello"}, "dua": [{"nama": "satu"}, {"nama": "dua"}]}}`},
		{"/root/dua", obj, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}, {"new": "hello"}]}}`},
		{"/root/dua", arr, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}, {"name": "john"}, {"name": "jane"}]}}`},
		{"/root/satu", arr, true, ""},
		{"/root/satu", str, true, ""},
		{"/root/empat", str, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}], "empat": "doe"}}`},
	}
	for i, v := range testData {
		ms, err := openapi.LoadFromBytes([]byte(source))
		if err != nil {
			t.Error("unable to parse source")
		}

		err = ms.AddPath(v.Path, v.Addition)
		if (err == nil) == v.ExpectedError {
			t.Errorf("expect error is %v, got `%v` on test data #%v (path: %v)", v.ExpectedError, err, i+1, v.Path)
		}

		b, _ := yaml.MarshalWithOptions(yaml.MapSlice(ms), yaml.JSON())
		str := strings.TrimSpace(string(b))
		if !v.ExpectedError && str != v.ExpectedResult {
			t.Errorf("test data #%v (path: %v)\n expect `%v`,\n    got `%v`", i+1, v.Path, v.ExpectedResult, str)
		}
	}
}

func TestMapSlice_SetPath(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	obj, arr, str := getAdditions(t)
	_, _, _ = obj, arr, str

	// Test Add to Path
	testData := []struct {
		Path           string
		Addition       any
		ExpectedError  bool
		ExpectedResult string
	}{
		{"/", obj, false, `{"new": "hello"}`},
		{"/", arr, true, ``},
		{"/", str, true, ``},
		{"/root", obj, false, `{"root": {"new": "hello"}}`},
		{"/root", arr, false, `{"root": [{"name": "john"}, {"name": "jane"}]}`},
		{"/root", str, false, `{"root": "doe"}`},
		{"/root/satu", obj, false, `{"root": {"satu": {"new": "hello"}, "dua": [{"nama": "satu"}, {"nama": "dua"}]}}`},
		{"/root/satu", arr, false, `{"root": {"satu": [{"name": "john"}, {"name": "jane"}], "dua": [{"nama": "satu"}, {"nama": "dua"}]}}`},
		{"/root/satu", str, false, `{"root": {"satu": "doe", "dua": [{"nama": "satu"}, {"nama": "dua"}]}}`},
		{"/root/dua/0", obj, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"new": "hello"}, {"nama": "dua"}]}}`},
		{"/root/dua/0", arr, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [[{"name": "john"}, {"name": "jane"}], {"nama": "dua"}]}}`},
		{"/root/dua/0", str, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": ["doe", {"nama": "dua"}]}}`},
		{"/next", obj, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}]}, "next": {"new": "hello"}}`},
		{"/next", arr, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}]}, "next": [{"name": "john"}, {"name": "jane"}]}`},
		{"/next", str, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}]}, "next": "doe"}`},
		{"/next/satu", obj, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}]}, "next": {"satu": {"new": "hello"}}}`},
		{"/next/satu", arr, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}]}, "next": {"satu": [{"name": "john"}, {"name": "jane"}]}}`},
		{"/next/satu", str, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}]}, "next": {"satu": "doe"}}`},
		{"/root/tiga", obj, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}], "tiga": {"new": "hello"}}}`},
		{"/root/tiga", arr, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}], "tiga": [{"name": "john"}, {"name": "jane"}]}}`},
		{"/root/tiga", str, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}], "tiga": "doe"}}`},
		{"/root/new/obj", str, false, `{"root": {"satu": {"nama": "John", "age": 17}, "dua": [{"nama": "satu"}, {"nama": "dua"}], "new": {"obj": "doe"}}}`},
	}
	for i, v := range testData {
		ms, err := openapi.LoadFromBytes([]byte(source))
		if err != nil {
			t.Error("unable to parse source")
		}

		err = ms.SetPath(v.Path, v.Addition)
		if (err == nil) == v.ExpectedError {
			t.Errorf("expect error is %v, got `%v` on test data #%v (path: %v)", v.ExpectedError, err, i+1, v.Path)
		}

		str := toJson(ms)
		if !v.ExpectedError && str != v.ExpectedResult {
			t.Errorf("test data #%v (path: %v)\n expect `%v`,\n    got `%v`", i+1, v.Path, v.ExpectedResult, str)
		}
	}
}

func toJson(ms openapi.MapSlice) string {
	b, _ := yaml.MarshalWithOptions(yaml.MapSlice(ms), yaml.JSON())
	return strings.TrimSpace(string(b))
}
