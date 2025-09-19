package main

import (
	"errors"
	"fmt"
	"lancache-dns-generator/formatters"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/adguardhome", formatters.GetAdGuardHomeList)

	err := http.ListenAndServe(":3333", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
