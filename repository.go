package main

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
)

type repository struct{}

func (repository) readLinesFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func (repository) writeCSVToFile(path string, contents [][]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	w.Write([]string{"h1", "meta_des", "title", "URL"})
	w.WriteAll(contents) // calls Flush internally

	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}

	return nil
}
