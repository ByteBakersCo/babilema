package history

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/ByteBakersCo/babilema/internal/config"
)

const (
	TimeFormat      string = time.RFC3339
	historyFileName string = ".babilema-history.toml"
	warningComment  string = "# This file is auto-generated by Babilema. Do not edit manually.\n\n"
)

type History struct {
	Data map[string]time.Time `toml:"history"`
}

func ParseHistoryFile(cfg config.Config) (map[string]time.Time, error) {
	history := History{
		Data: make(map[string]time.Time),
	}
	_, err := toml.DecodeFile(
		filepath.Join(cfg.OutputDir, historyFileName),
		&history,
	)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if errors.Is(err, os.ErrNotExist) {
		log.Println("Creating a new history file...")
	} else {
		log.Println("History file parsed.")
	}

	return history.Data, nil
}

func UpdateHistoryFile(history map[string]time.Time, cfg config.Config) error {
	var file *os.File

	file, err := os.OpenFile(
		filepath.Join(cfg.TempDir, historyFileName),
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(warningComment)
	if err != nil {
		return err
	}

	encoder := toml.NewEncoder(file)
	err = encoder.Encode(History{Data: history})
	if err != nil {
		return err
	}

	log.Println("History file updated.")

	return nil
}
