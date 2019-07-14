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

func NewDataLogger(csvPath string) (*DataLogger, error) {
	csvFile, err := os.Create(csvPath)
	if err != nil {
		return nil, err
	}
	return &DataLogger{
		logger: csv.NewWriter(csvFile),
		mutex:  &sync.Mutex{},
	}, nil
}
