package config

import (
	"testing"
	"time"
)

var ymlExample = `
store:
  book:
    - author: john
      price: 10
    - author: ken
      price: 12
  open: 24h
  rating: 4.5
  branch:
    malang:
      id: 1
    subang:
      id: 2
`

func TestUnmarshal(t *testing.T) {
	cfg, err := FromBytes([]byte(ymlExample))
	if err != nil {
		t.Error(err)
	}

	var d struct {
		Store struct {
			Book []struct {
				Author string
				Price  int
			}
			Open   time.Duration
			Rating float64
		}
	}
	err = cfg.Unmarshal(&d)
	if err != nil {
		t.Error(err)
	}

	if len(d.Store.Book) != 2 {
		t.Error("invalid length for book")
	}
	if d.Store.Book[0].Author != "john" || d.Store.Book[0].Price != 10 {
		t.Error("invalid value for book[0]")
	}
	if d.Store.Book[1].Author != "ken" || d.Store.Book[1].Price != 12 {
		t.Error("invalid value for book[1]")
	}
	if d.Store.Open != time.Duration(24*time.Hour) {
		t.Error("invalid value for open")
	}
	if d.Store.Rating != 4.5 {
		t.Error("invalid value for rating")
	}
}

func TestUnmarshalPath(t *testing.T) {
	cfg, err := FromBytes([]byte(ymlExample))
	if err != nil {
		t.Error(err)
	}

	var authors []string
	err = cfg.UnmarshalPath("$.store.book[*].author", &authors)
	if err != nil {
		t.Error(err)
	}
	if len(authors) != 2 || authors[0] != "john" || authors[1] != "ken" {
		t.Error("invalid authors, got:", authors)
	}

	var ids []int
	err = cfg.UnmarshalPath("$.store.branch..id", &ids)
	if err != nil {
		t.Error(err)
	}
	if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
		t.Error("invalid authors, got:", ids)
	}
}
