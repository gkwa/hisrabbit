package hisrabbit

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	LogFormat      string `long:"log-format" choice:"text" choice:"json" default:"text" description:"Log format"`
	Verbose        []bool `short:"v" long:"verbose" description:"Show verbose debug information, each -v bumps log level"`
	logLevel       slog.Level
	InputDataPath  string `short:"i" long:"input-path" description:"Path to the input JSON data file" required:"true"`
	OutputDataPath string `short:"o" long:"output-path" description:"Path to otuput the JSON data file to"`
}

var parser *flags.Parser

func Execute() int {
	parser = flags.NewParser(&opts, flags.HelpFlag)

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			parser.WriteHelp(os.Stdout)
			return 0
		}

		parser.WriteHelp(os.Stderr)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		return 1
	}

	if err := setLogLevel(); err != nil {
		slog.Error("error setting log level", "error", err)
		return 1
	}

	if err := setupLogger(); err != nil {
		slog.Error("error setting up logger", "error", err)
		return 1
	}

	if err := run(); err != nil {
		slog.Error("run failed", "error", err)
		return 1
	}

	return 0
}

func run() error {
	// Read the input JSON file
	data, err := os.ReadFile(opts.InputDataPath)
	if err != nil {
		return fmt.Errorf("error reading data.json: %v", err)
	}

	var records []Record
	err = json.Unmarshal(data, &records)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	records, err = Uniqueify(records)
	if err != nil {
		return fmt.Errorf("error uniqueifying data: %v", err)
	}

	// Marshal the sorted records back to JSON
	resultJSON, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	err = os.WriteFile(opts.OutputDataPath, resultJSON, 0o644)
	if err != nil {
		return fmt.Errorf("error writing to data1.json: %v", err)
	}

	return nil
}

func Uniqueify(records []Record) ([]Record, error) {
	// Create a map to store unique records based on the .path field
	uniqueRecords := make(map[string]Record)

	// Iterate over the records, keeping the record with the youngest indexed_at field
	for _, record := range records {
		// Retrieve the existing record from 'uniqueRecords' using the current record's 'Path' as the key
		existingRecord, exists := uniqueRecords[record.Path]

		// Check if the key doesn't exist in the map OR the current record has a younger 'IndexedAt' timestamp
		if !exists || record.IndexedAt.After(existingRecord.IndexedAt) {
			// If either condition is true, update 'uniqueRecords' with the current record using its 'Path' as the key
			uniqueRecords[record.Path] = record
		}
	}
	// Convert the map to a slice for sorting
	var sortedRecords []Record
	for _, record := range uniqueRecords {
		sortedRecords = append(sortedRecords, record)
	}

	// Sort the records based on the indexed_at field
	sort.Slice(sortedRecords, func(i, j int) bool {
		return sortedRecords[i].IndexedAt.Before(sortedRecords[j].IndexedAt)
	})

	return sortedRecords, nil
}
