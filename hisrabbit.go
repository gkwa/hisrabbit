package hisrabbit

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"time"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	LogFormat string `long:"log-format" choice:"text" choice:"json" default:"text" description:"Log format"`
	Verbose   []bool `short:"v" long:"verbose" description:"Show verbose debug information, each -v bumps log level"`
	logLevel  slog.Level

	DataPath string `short:"d" long:"data-path" description:"Path to the JSON data file" required:"true"`
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
	data, err := os.ReadFile(opts.DataPath)
	if err != nil {
		return fmt.Errorf("error reading data.json: %v", err)
	}

	var records []Record
	err = json.Unmarshal(data, &records)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}

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

	// Marshal the sorted records back to JSON
	resultJSON, err := json.MarshalIndent(sortedRecords, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	// Write the result to data1.json
	err = os.WriteFile("data1.json", resultJSON, 0o644)
	if err != nil {
		return fmt.Errorf("error writing to data1.json: %v", err)
	}

	return nil
}

// Record represents the structure of each record in the JSON data
type Record struct {
	BrowseURL string    `json:"browse_url"`
	CreatedAt time.Time `json:"created_at"`
	GitCommit string    `json:"git_commit"`
	GitURL    string    `json:"git_url"`
	IndexedAt time.Time `json:"indexed_at"`
	Path      string    `json:"path"`
	Release   string    `json:"release"`
	Subpath   string    `json:"subpath"`
	Version   string    `json:"version"`
}
