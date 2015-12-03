package bingo

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

/*
Application configuration.

 - Root: data folder path
 - Views: html templates folder
 - Static: static assets (js, css) folder
 - Log: log file
 - Stdout: when a log file is given, iset to true to still log on stdout
 - Verbosity: log verbosity mask
 - Port: webapp port
 - Depth: number of subfolders in data hierarchy (the more, the more folders, the fewer files per folder)
 - FloodThreshold: min delay (in seconds) between two posts for a single user
 - CleanThreshold: delete expired pasted from database once in that many seconds
*/
type Conf struct {
	Root           string `json:"root"`
	Views          string `json:"views"`
	Static         string `json:"static"`
	Log            string `json:"log"`
	Stdout         bool   `json:"stdout"`
	Verbosity      int    `json:"verbosity"`
	Port           int    `json:"port"`
	Depth          int    `json:"depth"`
	FloodThreshold int    `json:"floodThreshold"`
	CleanThreshold int    `json:"cleanThreshold"`
}

// Global configuration instance
var conf Conf

// Initializes configuration defaults
func init() {
	conf = Conf{
		Verbosity:      15, // All logs
		Port:           1337,
		Depth:          2,
		FloodThreshold: 10,
		CleanThreshold: 3600, // One hour
		Stdout:         false,
	}
}

// Load a configuration file
func (conf *Conf) load(path string) error {

	// Read file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Parse data
	if err := json.Unmarshal(data, &conf); err != nil {
		return err
	}

	// Clean folder paths
	conf.Root = filepath.Clean(conf.Root)
	conf.Views = filepath.Clean(conf.Views)
	conf.Static = filepath.Clean(conf.Static)
	if conf.Log != "" {
		conf.Log = filepath.Clean(conf.Log)
	}

	return nil
}
