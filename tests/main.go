package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"ruoCache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *ruoCache.Group {
	return createNewGroup("main")
}

func startCacheServer(addr string, addrs []string, ruo *ruoCache.Group) {
	peers := ruoCache.NewHttpPool(addr)
	peers.Set(addrs...)
	ruo.RegisterPeers(peers)
	log.Println("ruoCache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string) {
	http.Handle("/get", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			group := r.URL.Query().Get("group")
			if group == "" {
				group = "main"
			}
			view, err := getGroup(group).Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/application/json")
			w.Write(view.ByteSlice())
		}))

	http.Handle("/set", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			val := r.URL.Query().Get("val")
			group := r.URL.Query().Get("group")
			if group == "" {
				group = "main"
			}
			getGroup(group).Set(key, val)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{\"status\":0,\"code\":200, \"msg\":\"success\"}"))
		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "ruoCache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	ruo := createGroup()
	if api {
		go startAPIServer(apiAddr)
	}
	startCacheServer(addrMap[port], []string(addrs), ruo)
}

func createNewGroup(name string) *ruoCache.Group {
	return ruoCache.NewGroup(name, 2<<10, ruoCache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func getGroup(name string) *ruoCache.Group  {
	g := ruoCache.GetGroup(name)
	if g == nil {
		return createNewGroup(name)
	}
	return g
}
