package taxonomy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ovahol/ontology"
)

// Taxonomy is an alias for ontology.Taxonomy.
type Taxonomy = ontology.Taxonomy

// LoadFile reads and parses a taxonomy JSON file.
func LoadFile(path string) (*ontology.Taxonomy, error) {
	return ontology.LoadTaxonomyFile(path)
}

// LoadBytes parses taxonomy JSON bytes.
func LoadBytes(data []byte) (*ontology.Taxonomy, error) {
	return ontology.LoadTaxonomy(data)
}

// ResolvePath resolves a taxonomy file path using --taxonomy and --taxonomy-dir.
// If taxonomyPath is absolute or contains a separator and exists, it is used directly.
// If taxonomyDir is set, taxonomyPath is joined with it when taxonomyPath is relative.
// If taxonomyPath is empty but taxonomyDir is set, looks for *.json in dir and returns first.
func ResolvePath(taxonomyPath, taxonomyDir string) (string, error) {
	if taxonomyPath != "" {
		if filepath.IsAbs(taxonomyPath) {
			return taxonomyPath, nil
		}
		if taxonomyDir != "" {
			// if taxonomyPath is relative and taxonomyDir is set, join them
			joined := filepath.Join(taxonomyDir, taxonomyPath)
			if _, err := os.Stat(joined); err == nil {
				return joined, nil
			}
			// also try taxonomyPath as-is if joined doesn't exist
			if _, err := os.Stat(taxonomyPath); err == nil {
				return taxonomyPath, nil
			}
			return joined, nil
		}
		return taxonomyPath, nil
	}
	if taxonomyDir != "" {
		// try to find any json file in dir
		entries, err := os.ReadDir(taxonomyDir)
		if err != nil {
			return "", fmt.Errorf("taxonomy: read dir %s: %w", taxonomyDir, err)
		}
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
				return filepath.Join(taxonomyDir, e.Name()), nil
			}
		}
		return "", fmt.Errorf("taxonomy: no taxonomy file found in %s", taxonomyDir)
	}
	return "", fmt.Errorf("taxonomy: no taxonomy path or dir specified")
}
