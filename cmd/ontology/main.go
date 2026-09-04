package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ovahol/ontology"
)

func main() {
	var fromPath string
	var toPath string
	var taxonomyPath string
	var taxonomyDir string

	flag.StringVar(&fromPath, "from", "", "input workbook path")
	flag.StringVar(&toPath, "to", "", "output workbook path")
	flag.StringVar(&taxonomyPath, "taxonomy", "", "path to taxonomy JSON file")
	flag.StringVar(&taxonomyDir, "taxonomy-dir", "", "directory containing taxonomy JSON files (first *.json found is used)")
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

	// ontology is an engine: vendors supply their own taxonomy via --taxonomy
	// or --taxonomy-dir. If none is given, the embedded WHO/MeDevIS reference
	// vocabulary is used.
	var tax *ontology.Taxonomy
	resolved, err := resolveTaxonomyPath(taxonomyPath, taxonomyDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if resolved == "" {
		tax = ontology.DefaultTaxonomy()
	} else {
		tax, err = ontology.LoadTaxonomyFile(resolved)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	csvPath, err := ontology.NormalizeWorkbookWithTaxonomy(fromPath, toPath, tax)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("normalized workbook written to %s\n", toPath)
	fmt.Printf("api import csv written to %s\n", csvPath)
}

// resolveTaxonomyPath picks the taxonomy file to load: taxonomyPath if set
// (joined with taxonomyDir when relative and taxonomyDir is given), else the
// first *.json file in taxonomyDir, else "" (meaning: use the default).
func resolveTaxonomyPath(taxonomyPath, taxonomyDir string) (string, error) {
	if taxonomyPath != "" {
		if filepath.IsAbs(taxonomyPath) || taxonomyDir == "" {
			return taxonomyPath, nil
		}
		joined := filepath.Join(taxonomyDir, taxonomyPath)
		if _, err := os.Stat(joined); err == nil {
			return joined, nil
		}
		return taxonomyPath, nil
	}
	if taxonomyDir == "" {
		return "", nil
	}
	entries, err := os.ReadDir(taxonomyDir)
	if err != nil {
		return "", fmt.Errorf("read taxonomy dir %s: %w", taxonomyDir, err)
	}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			return filepath.Join(taxonomyDir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("no taxonomy file found in %s", taxonomyDir)
}
