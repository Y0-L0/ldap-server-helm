package sidecar

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const sentinelFile = ".initialized"

// seed loads LDIF files from seedDir into slapd.
// Skips seeding if a sentinel file exists at <dataDir>/.initialized.
func seed(backend Backend, seedDir, dataDir string) error {
	sentinel := filepath.Join(dataDir, sentinelFile)
	if _, err := os.Stat(sentinel); err == nil {
		slog.Info("sentinel exists, skipping seed", "path", sentinel)
		return nil
	}

	entries, err := loadLDIFDir(seedDir)
	if err != nil {
		return fmt.Errorf("loading LDIF files: %w", err)
	}

	for _, entry := range entries {
		slog.Info("seeding entry", "dn", entry.dn)
		if err := backend.Add(entry.dn, entry.attrs); err != nil {
			return fmt.Errorf("adding %s: %w", entry.dn, err)
		}
	}

	if err := os.WriteFile(
		sentinel,
		[]byte("initialized\n"),
		0o600,
	); err != nil {
		return fmt.Errorf("writing sentinel: %w", err)
	}

	slog.Info("seed complete", "entries", len(entries))
	return nil
}

type ldifEntry struct {
	dn    string
	attrs map[string][]string
}

func loadLDIFDir(dir string) ([]ldifEntry, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.ldif"))
	if err != nil {
		return nil, fmt.Errorf("globbing LDIF files: %w", err)
	}
	sort.Strings(files)

	var all []ldifEntry
	for _, f := range files {
		entries, err := parseLDIF(f)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", f, err)
		}
		all = append(all, entries...)
	}

	return all, nil
}

func parseLDIF(path string) ([]ldifEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []ldifEntry
	var current *ldifEntry

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if current != nil {
				entries = append(entries, *current)
				current = nil
			}
			continue
		}

		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			continue
		}

		if key == "dn" {
			current = &ldifEntry{
				dn:    value,
				attrs: make(map[string][]string),
			}
			continue
		}

		if current != nil {
			current.attrs[key] = append(current.attrs[key], value)
		}
	}

	if current != nil {
		entries = append(entries, *current)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
