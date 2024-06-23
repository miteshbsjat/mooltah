package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"encoding/json"

	"text/template"

	_ "./embedded_binaries"
	"github.com/BurntSushi/toml"
	"github.com/alexflint/go-arg"
	"github.com/noirbizarre/gonja"
	"gopkg.in/yaml.v2"
)

//go:embed binary_arm64
var binaryArm64 embed.FS

//go:embed binary_amd64
var binaryAmd64 embed.FS

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
	var binaryToExecute []byte
	arch := runtime.GOARCH

	switch arch {
	case "arm64":
		data, err := binaryArm64.ReadFile("binary_arm64")
		if err != nil {
			fmt.Println(err)
			return
		}
		binaryToExecute = data
	case "amd64":
		data, err := binaryAmd64.ReadFile("binary_amd64")
		if err != nil {
			fmt.Println(err)
			return
		}
		binaryToExecute = data
	default:
		fmt.Printf("Unsupported architecture: %s\n", arch)
		return
	}

	// Write the binary to a temporary file and execute it
	tmpfile, err := os.CreateTemp("", "binary")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.Remove(tmpfile.Name()) // Remove the temp file after execution

	if _, err = tmpfile.Write(binaryToExecute); err != nil {
		fmt.Println(err)
		return
	}
	if err := tmpfile.Close(); err != nil {
		fmt.Println(err)
		return
	}

	cmd := exec.Command(tmpfile.Name()) // Execute the binary
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to execute binary: %s\n", string(output))
	} else {
		fmt.Println("Binary executed successfully")
	}

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
}

func runMooltah(arguments args) error {
	files := arguments.Variable
	outputFile := arguments.Output
	templateFile := arguments.InputTemplateFile
	slog.Info(templateFile)

	var result map[string]interface{}

	result, err := processFiles(files)
	if err != nil {
		slog.Error("Failed %v", err)
		return err
	}

	fmt.Println(result)
	output, err := renderTemplateFile(templateFile, &result, outputFile)
	if err != nil {
		slog.Error("Failed %v", err)
		return err
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
