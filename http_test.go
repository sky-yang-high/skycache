package skycache

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	p := NewHTTPPool("localhost:xxx")
	NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	if groups["scores"] == nil {
		t.Fatal("newGroup error")
	}

	server := httptest.NewServer(p)
	defer server.Close()

	//t.Log(server.URL)

	base := server.URL + defaultBasePath
	req1, err := http.NewRequest("GET", base+"scores/Tom", nil)
	if err != nil {
		t.Fatalf("creating request error: %s\n", err)
	}

	//t.Log(req1.URL)

	client := server.Client()

	resp1, err := client.Do(req1)
	if err != nil || resp1.StatusCode != http.StatusOK {
		t.Fatalf("do request error: %d\n", resp1.StatusCode)
	}

	buf, err := io.ReadAll(resp1.Body)
	if err != nil {
		t.Fatalf("reading from response error: %s\n", err)
	}
	if string(buf) != db["Tom"] {
		t.Errorf("result error,expected %s, got %s\n", db["Tom"], string(buf))
	}
}

func TestMain(t *testing.T) {
	NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := NewHTTPPool(addr)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
