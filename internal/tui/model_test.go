package tui

import (
	"testing"
	"time"

	"github.com/b00y0h/wakadash/internal/archive"
	"github.com/b00y0h/wakadash/internal/datasource"
)

func TestNewModel_WithDataSource(t *testing.T) {
	// Test that NewModel accepts a DataSource
	fetcher := archive.New("owner/repo")
	ds := datasource.New(nil, fetcher)
	m := NewModel(nil, "last_7_days", 60*time.Second, ds)

	if m.dataSource == nil {
		t.Error("expected dataSource to be set")
	}
}

func TestNewModel_WithNilDataSource(t *testing.T) {
	// Test that NewModel handles nil dataSource gracefully
	m := NewModel(nil, "last_7_days", 60*time.Second, nil)

	if m.dataSource != nil {
		t.Error("expected dataSource to be nil")
	}
}

func TestArchiveFetchedMsg_NilData(t *testing.T) {
	// Test that archiveFetchedMsg with nil data doesn't cause error
	m := NewModel(nil, "last_7_days", 60*time.Second, nil)

	msg := archiveFetchedMsg{data: nil, date: "2026-02-24"}
	newModel, _ := m.Update(msg)

	// Cast back to Model to check state
	model := newModel.(Model)
	if model.archiveData != nil {
		t.Error("expected archiveData to be nil after nil msg")
	}
	// No error state should be set
	if model.err != nil {
		t.Error("expected no error from nil archive data")
	}
}
