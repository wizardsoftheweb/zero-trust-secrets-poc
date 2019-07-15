package main

import (
	"encoding/csv"
	"os"
	"sync"
)

type DataLogger struct {
	logger *csv.Writer
	mutex  *sync.Mutex
}

func (d *DataLogger) Log(row []string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	err := d.logger.Write(row)
	return err
}

func (d *DataLogger) Flush() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.logger.Flush()
}

// Updated with commentary from reddit
// https://www.reddit.com/r/golang/comments/46fdef/help_appending_a_slice_to_next_row_in_csv_file/
func NewDataLogger(csvPath string) (*DataLogger, error) {
	csvFile, err := os.OpenFile(
		csvPath,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}
	return &DataLogger{
		logger: csv.NewWriter(csvFile),
		mutex:  &sync.Mutex{},
	}, nil
}
