package ruoCache

import (
	"fmt"
	"log"
	"testing"
)

var db = map[string]string{
	"eddie":"678",
	"alreadyshow":"67812",
	"tom":"6781",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	ruo := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	for k, v := range db {
		if view, err := ruo.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value of %s", k)
		} // load from callback function
		if _, err := ruo.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := ruo.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
