package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"encoding/hex"
	"encoding/json"

	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/alexflint/go-arg"
	"github.com/noirbizarre/gonja"
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

func writeOutput(data string, filename string) error {
	err := os.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		return err
	}
	return nil
}

// createYAMLFile function takes a map and a file name, and writes the map to the file in YAML format
func createYAMLFile(data *map[string]interface{}, filename string) error {
	// Marshal the map into YAML format
	yamlData, err := yaml.Marshal(&data)
	if err != nil {
		return fmt.Errorf("error marshaling data to YAML: %w", err)
	}

	// Write the YAML data to a file
	err = os.WriteFile(filename, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("error writing YAML to file: %w", err)
	}

	return nil
}

func renderTemplateFileJ2(templateFile string, data *map[string]interface{}, outputFile string) (string, error) {
	tpl := gonja.Must(gonja.FromFile(templateFile))

	// Execute the template
	output, err := tpl.Execute(*data)
	if err != nil {
		fmt.Println(err)
		// slog.Error("Template Rendition " + string(err))
		return "", err
	}

	// Print or use the output as needed
	slog.Debug(output)

	return output, nil
}

func renderTemplateFileMJ(templateFile string, data *map[string]interface{}, outputFile string) (string, error) {
	arch := runtime.GOARCH

	// Get the path of the current Go binary
	executablePath, err := os.Executable()
	if err != nil {
		return "error getting current executable path", err
	}
	executableDir := filepath.Dir(executablePath)

	// Get the name of the current Go binary
	executableName := fmt.Sprintf("minijinja-cli_%s", arch)

	// Construct the full path of the executable to run
	executableFullPath := filepath.Join(executableDir, executableName)

	// Check if the executable exists and is a file
	if _, err := os.Stat(executableFullPath); os.IsNotExist(err) {
		return fmt.Sprintf("executable not found: %s", executableFullPath), err
	} else if err != nil {
		return fmt.Sprintf("error checking executable: %w", err), err
	}

	// Marshal the map into JSON format
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshaling data to JSON: %w", err)
	}
	hash := sha256.Sum256(jsonData)
	hashStr := hex.EncodeToString(hash[0:5])
	// create dataFile yaml from data map
	dataFile := fmt.Sprintf("/tmp/mooltah-%s.yaml", hashStr)
	if err := createYAMLFile(data, dataFile); err != nil {
		return fmt.Sprintf("Failed to create dataFile %s", &dataFile), err
	}
	slog.Info(dataFile)
	slog.Info(executableFullPath)
	// defer os.Remove(dataFile)

	// cmd := exec.Command("/bin/bash", "-c", executableFullPath, "-o", outputFile, templateFile, dataFile)
	cmd := exec.Command(executableFullPath, "-o", outputFile, templateFile, dataFile)
	slog.Info(cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to execute binary: %s", string(output)), err
	} else {
		fmt.Println("Binary executed successfully")
	}
	fmt.Println(string(output))

	return "", nil
}

func renderTemplateFile(templateFile string, data *map[string]interface{}, outputFile string) (string, error) {
	tpl, err := template.ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}

	ofh, err := os.Create(outputFile)
	defer ofh.Close()
	// writer := bufio.NewWriter(ofh)

	// Execute the template
	err = tpl.Execute(ofh, *data)
	if err != nil {
		fmt.Println(err)
		// slog.Error("Template Rendition " + string(err))
		return "", err
	}

	// Print or use the output as needed
	// slog.Debug(output)

	return "", nil
}

type args struct {
	Variable          []string `arg:"-v,--variable,separate" help:"Read variables from YAML, JSON, TOML, and/or Key=Value Files"`
	InputTemplateFile string   `arg:"positional" help:"Template File which will be rendered to OUTPUT"`
	Output            string   `arg:"-o,--output" help:"Output file which will have rendition of input template file"`
	TemplateType      string   `arg:"-t,--template-type" default:"jinja2"`
}

func runMooltah(arguments args) error {
	files := arguments.Variable
	outputFile := arguments.Output
	templateFile := arguments.InputTemplateFile
	templateType := arguments.TemplateType
	slog.Info(templateFile)

	var result map[string]interface{}

	result, err := processFiles(files)
	if err != nil {
		fmt.Printf("Failed %s; %w", result, err)
		return err
	}

	fmt.Println(result)
	var output string
	var errt error
	if templateType == "jinja2" {
		output, errt = renderTemplateFileMJ(templateFile, &result, outputFile)
	} else {
		output, errt = renderTemplateFile(templateFile, &result, outputFile)
	}
	if errt != nil {
		fmt.Printf("Failed %s; %w", output, errt)
		return errt
	}

	if output != "" {
		slog.Info("Aaya")
		err = writeOutput(output, outputFile)
		if err != nil {
			slog.Error("Failed %v", err)
			return err
		}
	}
	return nil
}

func (args) Version() string {
	return "mooltah 1.0.0"
}

func (args) Description() string {
	return "this program renders input template to output with the configurations given by --variable files"
}

func main() {
	var arguments args
	arg.MustParse(&arguments)
	if arguments.InputTemplateFile == "" {
		slog.Error("Please provide InputTemplateFile")
		os.Exit(1)
	}
	if arguments.Output == "" {
		slog.Error("Please provide Output file")
		os.Exit(1)
	}
	err := runMooltah(arguments)
	if err != nil {
		slog.Error("Error ", err)
		os.Exit(2)
	}
}
