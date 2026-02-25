package tui

import (
	"testing"
	"time"

	"github.com/b00y0h/wakadash/internal/archive"
)

func TestNewModel_WithArchiveFetcher(t *testing.T) {
	// Test that NewModel accepts an archive fetcher
	fetcher := archive.New("owner/repo")
	m := NewModel(nil, "last_7_days", 60*time.Second, fetcher)

	if m.archiveFetcher == nil {
		t.Error("expected archiveFetcher to be set")
	}
}

func TestNewModel_WithNilArchiveFetcher(t *testing.T) {
	// Test that NewModel handles nil fetcher gracefully
	m := NewModel(nil, "last_7_days", 60*time.Second, nil)

	if m.archiveFetcher != nil {
		t.Error("expected archiveFetcher to be nil")
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
