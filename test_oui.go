//go:build ignore

package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func main() {
	data, _ := os.ReadFile("internal/scan/oui.csv")
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		fmt.Println("CSV parse error:", err)
		return
	}
	fmt.Println("Total records:", len(records))
	found := false
	for _, rec := range records {
		if len(rec) >= 3 && strings.ToLower(rec[1]) == "3c6d66" {
			fmt.Println("Found:", rec[1], rec[2])
			found = true
		}
	}
	if !found {
		fmt.Println("3C6D66 NOT FOUND")
	}
}
