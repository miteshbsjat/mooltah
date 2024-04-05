package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"encoding/json"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func parseKVFile(data []byte, result *map[string]interface{}) error {
	if data == nil || len(data) == 0 {
		return errors.New("data is empty or null")
	}
	*result = make(map[string]interface{})
	reader := bytes.NewBuffer(data)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		trimmedLine := bytes.TrimSpace(line)
		if len(trimmedLine) == 0 || trimmedLine[0] == '#' {
			continue // skip empty or commented lines
		}
		keyValue := bytes.SplitN(trimmedLine, []byte("="), 2)
		if len(keyValue) != 2 {
			return errors.New("invalid key=value format")
		}
		slog.Debug("key = " + string(bytes.TrimSpace(keyValue[0])))
		(*result)[string(bytes.TrimSpace(keyValue[0]))] = string(bytes.TrimSpace(keyValue[1]))
	}
	return nil
}

// mergeMaps merges two maps into one.
func mergeMaps(m1, m2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m1 {
		result[k] = v
	}

	for k, v := range m2 {
		result[k] = v
	}

	return result
}

func processFiles(files []string) (map[string]interface{}, error) {

	result := make(map[string]interface{})

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			slog.Error("Error reading file %s: %s\n", file, err)
			return result, err
		}

		ext := filepath.Ext(file)
		slog.Debug("ext = " + ext)
		switch ext {
		case ".yaml", ".yml":
			var yamlData map[string]interface{}
			err = yaml.Unmarshal(data, &yamlData)
			if err != nil {
				slog.Error("Error unmarshalling YAML file %s: %s\n", file, err)
				return result, err
			}
			result = mergeMaps(result, yamlData)
		case ".json":
			var jsonData map[string]interface{}
			err = json.Unmarshal(data, &jsonData)
			if err != nil {
				slog.Error("Error unmarshalling JSON file %s: %s\n", file, err)
				return result, err
			}
			result = mergeMaps(result, jsonData)
		case ".toml":
			var tomlData map[string]interface{}
			_, err = toml.Decode(string(data), &tomlData)
			if err != nil {
				slog.Error("Error unmarshalling TOML file %s: %s\n", file, err)
				return result, err
			}
			result = mergeMaps(result, tomlData)
		case ".kv", ".txt":
			var kvData map[string]interface{}
			err = parseKVFile(data, &kvData)
			if err != nil {
				slog.Error("Error unmarshalling Key-Value file %s: %s\n", file, err)
				return result, err
			}
			result = mergeMaps(result, kvData)
		default:
			slog.Warn("Unsupported file format for file %s\n", file)
		}
	}

	return result, nil
}

var rootCmd = &cobra.Command{
	Use:   "mooltah",
	Short: "A command-line tool for processing YAML, JSON, and TOML files",
	Run:   runMooltah,
}

func runMooltah(cmd *cobra.Command, args []string) {
	files := viper.GetStringSlice("files")

	var result map[string]interface{}

	result, err := processFiles(files)
	if err != nil {
		slog.Error("Failed %v", err)
		return
	}

	fmt.Println(result)
}

func main() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringSliceP("files", "v", []string{}, "List of files to process")
	viper.BindPFlag("files", rootCmd.Flags().Lookup("files"))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}
