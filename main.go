package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/noirbizarre/gonja"
	"gopkg.in/yaml.v2"
)

func main() {
	// Define command-line flags
	templateFile := flag.String("template", "", "Path to the Jinja2 template file")
	dataFile := flag.String("data", "", "Path to the YAML data file")
	flag.Parse()

	// Check if both template and data files are provided
	if *templateFile == "" || *dataFile == "" {
		fmt.Println("Both template and data files are required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Read template file
	// templateContent, err := ioutil.ReadFile(*templateFile)
	templateContent, err := ioutil.ReadFile(*templateFile)
	if err != nil {
		fmt.Printf("Error reading template file: %s\n", err)
		os.Exit(1)
	}

	// Read data file
	dataContent, err := ioutil.ReadFile(*dataFile)
	if err != nil {
		fmt.Printf("Error reading data file: %s\n", err)
		os.Exit(1)
	}

	// Parse YAML data
	var data map[string]interface{}
	if err := yaml.Unmarshal(dataContent, &data); err != nil {
		fmt.Printf("Error parsing YAML data: %s\n", err)
		os.Exit(1)
	}

	// Compile the template first (i. e. creating the AST)
	tpl, err := gonja.FromString(string(templateContent))
	if err != nil {
		panic(err)
	}
	// Now you can render the template with the given
	// gonja.Context how often you want to.
	// out, err := tpl.Execute(gonja.Context{"name": "axel"})
	out, err := tpl.Execute(data)
	if err != nil {
		panic(err)
	}

	// Print the rendered output
	fmt.Println(out)
}
