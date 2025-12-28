package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v3"
)

// Formatter defines the interface for output formatting
type Formatter interface {
	Format(data interface{}) error
}

// NewFormatter creates a new formatter based on the format type
func NewFormatter(format string, writer io.Writer) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{writer: writer}
	case "yaml":
		return &YAMLFormatter{writer: writer}
	case "table":
		fallthrough
	default:
		return &TableFormatter{writer: writer}
	}
}

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	writer io.Writer
}

// Format formats the data as JSON
func (f *JSONFormatter) Format(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// YAMLFormatter formats output as YAML
type YAMLFormatter struct {
	writer io.Writer
}

// Format formats the data as YAML
func (f *YAMLFormatter) Format(data interface{}) error {
	encoder := yaml.NewEncoder(f.writer)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(data)
}

// TableFormatter formats output as a table
type TableFormatter struct {
	writer io.Writer
}

// Format formats the data as a table
func (f *TableFormatter) Format(data interface{}) error {
	// Handle nil data
	if data == nil {
		return nil
	}

	// Check if data is a slice
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Slice {
		if val.Len() == 0 {
			fmt.Fprintln(f.writer, "No resources found")
			return nil
		}
		return f.formatSlice(val)
	}

	// Single object
	return f.formatSingle(data)
}

func (f *TableFormatter) formatSlice(val reflect.Value) error {
	if val.Len() == 0 {
		return nil
	}

	// Get the first element to determine headers
	first := val.Index(0)
	headers, rows := f.extractTableData(first)

	// Extract data for all elements
	allRows := [][]string{rows}
	for i := 1; i < val.Len(); i++ {
		_, row := f.extractTableData(val.Index(i))
		allRows = append(allRows, row)
	}

	// Create tabwriter
	w := tabwriter.NewWriter(f.writer, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Print rows
	for _, row := range allRows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return nil
}

func (f *TableFormatter) formatSingle(data interface{}) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		fmt.Fprintf(f.writer, "%v\n", data)
		return nil
	}

	typ := val.Type()

	// Create tabwriter
	w := tabwriter.NewWriter(f.writer, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "FIELD\tVALUE")

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldName := f.formatHeader(field.Name)
		fieldValue := f.formatValue(value)

		fmt.Fprintf(w, "%s\t%s\n", fieldName, fieldValue)
	}

	return nil
}

func (f *TableFormatter) extractTableData(val reflect.Value) ([]string, []string) {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var headers []string
	var row []string

	typ := val.Type()

	// Define key fields for each resource type
	keyFields := f.getKeyFields(typ.Name())

	for _, fieldName := range keyFields {
		field, found := typ.FieldByName(fieldName)
		if !found {
			continue
		}

		if !field.IsExported() {
			continue
		}

		headers = append(headers, f.formatHeader(fieldName))

		fieldVal := val.FieldByName(fieldName)
		row = append(row, f.formatValue(fieldVal))
	}

	return headers, row
}

func (f *TableFormatter) getKeyFields(typeName string) []string {
	switch typeName {
	case "Namespace":
		return []string{"ID", "Name", "PolicyPriority", "MergeStrategy", "CreatedAt"}
	case "Group":
		return []string{"ID", "Name", "NamespaceID", "Priority", "CreatedAt"}
	case "Target":
		return []string{"ID", "Hostname", "IPAddress", "Status", "CreatedAt"}
	case "BootstrapToken":
		return []string{"ID", "Name", "MaxUses", "Uses", "IsValid", "ExpiresAt"}
	default:
		// Default: return all fields
		return nil
	}
}

func (f *TableFormatter) formatHeader(name string) string {
	// Convert CamelCase to UPPER_CASE with spaces
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		if r >= 'a' && r <= 'z' {
			result.WriteRune(r - 32) // Convert to uppercase
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (f *TableFormatter) formatValue(val reflect.Value) string {
	if !val.IsValid() {
		return ""
	}

	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return ""
		}
		return f.formatValue(val.Elem())
	case reflect.Slice, reflect.Array:
		if val.Len() == 0 {
			return ""
		}
		// Format slice as comma-separated values
		var items []string
		for i := 0; i < val.Len() && i < 3; i++ { // Limit to first 3 items
			items = append(items, f.formatValue(val.Index(i)))
		}
		result := strings.Join(items, ", ")
		if val.Len() > 3 {
			result += "..."
		}
		return result
	case reflect.Map:
		if val.Len() == 0 {
			return ""
		}
		return fmt.Sprintf("(%d items)", val.Len())
	case reflect.Struct:
		// Handle time.Time specially
		if val.Type().String() == "time.Time" {
			t := val.Interface().(time.Time)
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02 15:04:05")
		}
		// For other structs, try to get a meaningful representation
		if val.Type().Name() != "" {
			return fmt.Sprintf("<%s>", val.Type().Name())
		}
		return "<struct>"
	case reflect.Bool:
		if val.Bool() {
			return "true"
		}
		return "false"
	case reflect.String:
		s := val.String()
		// Truncate long strings
		if len(s) > 50 {
			return s[:47] + "..."
		}
		return s
	default:
		return fmt.Sprintf("%v", val.Interface())
	}
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Fprintf(os.Stdout, "✓ %s\n", message)
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Fprintf(os.Stderr, "✗ %s\n", message)
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	fmt.Fprintf(os.Stdout, "ℹ %s\n", message)
}
