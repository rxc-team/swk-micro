package csvx

import (
	"encoding/csv"
	"io"
	"sync"
)

type SyncWriter struct {
	m      sync.Mutex
	Writer *csv.Writer
}

func NewSyncWriter(f io.Writer) *SyncWriter {
	w := csv.NewWriter(f)
	return &SyncWriter{Writer: w, m: sync.Mutex{}}

}

func (w *SyncWriter) Write(row []string) error {
	w.m.Lock()
	defer w.m.Unlock()
	err := w.Writer.Write(row)
	if err != nil {
		return err
	}

	return nil
}

func (w *SyncWriter) WriteAll(table [][]string) error {
	w.m.Lock()
	defer w.m.Unlock()
	err := w.Writer.WriteAll(table)
	if err != nil {
		return err
	}

	return nil
}

func (w *SyncWriter) Flush() {
	w.m.Lock()
	defer w.m.Unlock()
	w.Writer.Flush()
}
