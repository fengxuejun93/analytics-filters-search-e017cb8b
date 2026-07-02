package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	store := NewStore()
	handler := NewHandler(store)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	addr := ":8080"
	fmt.Printf("二手货品置换平台已启动 → http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
