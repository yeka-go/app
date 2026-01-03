package merger_test

import (
	"testing"

	"github.com/yeka-go/app/cmd/goapp/internal/openapi/merger"
)

func TestCombinePath(t *testing.T) {
	testData := []struct {
		Base   string
		Add    string
		Sub    string
		Expect string
	}{
		{
			Base:   "/",
			Add:    "/paths/~1hello/$ref",
			Sub:    "/",
			Expect: "/paths/~1hello",
		},
		{
			Base:   "/paths/~1hello",
			Add:    "/paths/~1hello/get/responses/200/content/application~1json/schema/properties/name/$ref",
			Sub:    "/paths/~1hello",
			Expect: "/paths/~1hello/get/responses/200/content/application~1json/schema/properties/name",
		},
		{
			Base:   "/paths/~1hello",
			Add:    "/get/$ref",
			Sub:    "/",
			Expect: "/paths/~1hello/get",
		},
	}
	for i, v := range testData {
		res := merger.CombinePath(v.Base, v.Add, v.Sub)
		if res != v.Expect {
			t.Errorf("\nscenario #%v, expect %v, got %v", i+1, v.Expect, res)
		}
	}
}
