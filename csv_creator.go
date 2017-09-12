package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func main() {
	var records [][]string
	for i := 1; i < 3000; i++ {
		record := []string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("GCPUG Tシャツ No %d", i),
			fmt.Sprintf("%d", 1000+i),
		}
		records = append(records, record)
	}

	file, err := os.Create(`/tmp/test.csv`)
	if err != nil {
		fmt.Errorf("%v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.WriteAll(records)
}
