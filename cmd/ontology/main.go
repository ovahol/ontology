package main

import (
	"fmt"
	"os"

	"github.com/ovahol/ontology"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <input.xlsx> <output.xlsx>\n", os.Args[0])
		os.Exit(2)
	}
	inputPath := os.Args[1]
	outputPath := os.Args[2]
	csvPath, err := ontology.NormalizeWorkbook(inputPath, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("normalized workbook written to %s\n", outputPath)
	fmt.Printf("api import csv written to %s\n", csvPath)

	if err := validateCompliance(); err != nil {
		fmt.Fprintf(os.Stderr, "validation warning: %v\n", err)
	}
}

func validateCompliance() error {
	if len(ontology.OvaholDeviceTypes) != 8 {
		return fmt.Errorf("expected 8 Ovahol device types, got %d", len(ontology.OvaholDeviceTypes))
	}
	if len(ontology.DeviceFunctions) != 9 {
		return fmt.Errorf("expected 9 device functions, got %d", len(ontology.DeviceFunctions))
	}
	if len(ontology.DeviceApplicationRisks) != 5 {
		return fmt.Errorf("expected 5 application risks, got %d", len(ontology.DeviceApplicationRisks))
	}
	return nil
}
