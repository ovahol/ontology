package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ovahol/ontology"
	"github.com/ovahol/ontology/taxonomy"
)

func main() {
	var fromPath string
	var toPath string
	var taxonomyPath string
	var taxonomyDir string

	flag.StringVar(&fromPath, "from", "", "input workbook path")
	flag.StringVar(&toPath, "to", "", "output workbook path")
	flag.StringVar(&taxonomyPath, "taxonomy", "", "path to taxonomy JSON file")
	flag.StringVar(&taxonomyDir, "taxonomy-dir", "", "directory containing taxonomy JSON files")
	flag.Parse()

	// positional compat: ontology <input.xlsx> <output.xlsx>
	args := flag.Args()
	if fromPath == "" && len(args) > 0 {
		fromPath = args[0]
		args = args[1:]
	}
	if toPath == "" && len(args) > 0 {
		toPath = args[0]
		args = args[1:]
	}

	if fromPath == "" || toPath == "" {
		fmt.Fprintf(os.Stderr, "usage: %s [--from <input.xlsx>] [--to <output.xlsx>] [--taxonomy <file>] [--taxonomy-dir <dir>] [<input.xlsx> <output.xlsx>]\n", os.Args[0])
		flag.Usage()
		os.Exit(2)
	}

	// Vocab-less: vendor must supply taxonomy via --taxonomy or --taxonomy-dir.
	if taxonomyPath == "" && taxonomyDir == "" {
		fmt.Fprintf(os.Stderr, "error: ontology is vocab-less — supply --taxonomy <file> or --taxonomy-dir <dir>\n")
		fmt.Fprintf(os.Stderr, "  example: %s --taxonomy examples/taxonomies/ovahol.json %s %s\n", os.Args[0], fromPath, toPath)
		os.Exit(2)
	}
	resolved := taxonomyPath
	if taxonomyPath != "" && taxonomyDir != "" {
		if _, err := os.Stat(taxonomyPath); os.IsNotExist(err) {
			candidate := taxonomyDir + "/" + taxonomyPath
			if _, err2 := os.Stat(candidate); err2 == nil {
				resolved = candidate
			} else if p, err3 := taxonomy.ResolvePath(taxonomyPath, taxonomyDir); err3 == nil {
				resolved = p
			}
		}
	} else if taxonomyPath == "" && taxonomyDir != "" {
		p, err := taxonomy.ResolvePath("", taxonomyDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		resolved = p
	}
	t, err := taxonomy.LoadFile(resolved)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Vocab-less workbook — taxonomy is required.
	var csvPath string
	csvPath, err = ontology.NormalizeWorkbookWithTaxonomy(fromPath, toPath, t)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("normalized workbook written to %s\n", toPath)
	fmt.Printf("api import csv written to %s\n", csvPath)

	if err := validateCompliance(); err != nil {
		fmt.Fprintf(os.Stderr, "validation warning: %v\n", err)
	}
}

func validateCompliance() error {
	// Vocab-less: ontology no longer validates built-in counts. Vendors validate
	// via Taxonomy.Validate() on their own file. Keep stub for compat.
	return nil
}
