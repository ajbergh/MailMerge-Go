/*
File Service - Contact File Import

This service imports contact lists from CSV and Excel files into the shared
Contact model. It validates structure, protects import resource usage, and keeps
all source columns available to the merge-field schema.

Supported Formats:
  - CSV (.csv) - Comma-separated values with header row
  - Excel (.xlsx) - Microsoft Excel 2007+ format (first worksheet)

Column Detection:
  - Case-insensitive column name matching
  - Supports variations: "FirstName", "First Name", "first_name"
  - Email column is required; all other columns are optional

Error handling:
  - Returns row-level errors and warnings without discarding valid contacts
  - Reports duplicate email groups for preflight policy resolution
  - Validates addresses with the shared net/mail-based email package
*/
package services

import (
	"MailMergeApp/backend/email"
	"MailMergeApp/backend/models"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// Import safeguards (M4.20): bound resource usage so a malformed or hostile file
// cannot exhaust memory.
const (
	MaxImportFileBytes = 50 * 1024 * 1024 // 50 MB
	MaxImportRows      = 100000
	MaxImportColumns   = 512
)

// FileService handles importing contacts from CSV and Excel files.
// Stateless service - each parse operation is independent.
type FileService struct{}

// NewFileService creates a new FileService instance.
func NewFileService() *FileService {
	return &FileService{}
}

// stripBOM removes a leading UTF-8 byte-order mark from a header cell.
func stripBOM(s string) string {
	return strings.TrimPrefix(s, "\xef\xbb\xbf")
}

// ParseContactFile parses a CSV or XLSX contact file, returns normalized headers
// and row diagnostics, and records duplicate recipient groups for preflight.
//
// Parameters:
//   - filePath: Absolute path to the CSV or Excel file
//
// Returns:
//   - *models.ParseResult: Contains parsed contacts, headers, and any row-level errors
//   - error: Fatal errors that prevented parsing (file not found, etc.)
func (fs *FileService) ParseContactFile(filePath string) (*models.ParseResult, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	var result *models.ParseResult
	var err error

	switch ext {
	case ".csv":
		result, err = fs.parseCSV(filePath)
	case ".xlsx":
		result, err = fs.parseExcel(filePath)
	default:
		return nil, fmt.Errorf("unsupported file format: %s. Supported formats: .csv, .xlsx", ext)
	}

	if err != nil {
		return nil, err
	}

	// Detect duplicates after parsing
	fs.detectDuplicates(result)

	return result, nil
}

// parseCSV parses a CSV file with a required header row and retains every column
// as a contact field.
func (fs *FileService) parseCSV(filePath string) (*models.ParseResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)

	// Read header row
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Strip a UTF-8 BOM from the first header cell if present.
	if len(header) > 0 {
		header[0] = stripBOM(header[0])
	}
	if len(header) > MaxImportColumns {
		return nil, fmt.Errorf("too many columns (%d); limit is %d", len(header), MaxImportColumns)
	}

	normalizedHeaders, headerWarnings := normalizeHeaders(header)

	// Find column indices (case-insensitive)
	colIndices := fs.findColumnIndices(header)
	if colIndices["email"] == -1 {
		return nil, fmt.Errorf("required column 'Email' not found in CSV")
	}

	result := &models.ParseResult{
		Contacts: []models.Contact{},
		Errors:   []string{},
		Warnings: append([]string{}, headerWarnings...),
		Headers:  normalizedHeaders,
	}

	rowNum := 1 // Start after header
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to read row: %v", rowNum+1, err))
			rowNum++
			continue
		}
		if len(result.Contacts) >= MaxImportRows {
			result.Warnings = append(result.Warnings, fmt.Sprintf("import truncated at %d rows", MaxImportRows))
			break
		}
		if isBlankRow(row) {
			rowNum++
			continue
		}

		contact, err := fs.extractContact(row, colIndices, normalizedHeaders, rowNum+1)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
		} else if contact.Email != "" {
			result.Contacts = append(result.Contacts, contact)
		}
		rowNum++
	}

	result.Total = len(result.Contacts)
	return result, nil
}

// normalizeHeaders trims headers and reports duplicate (case-insensitive) column
// names as warnings.
func normalizeHeaders(raw []string) (headers []string, warnings []string) {
	headers = make([]string, len(raw))
	seen := map[string]bool{}
	for i, h := range raw {
		h = strings.TrimSpace(h)
		headers[i] = h
		if h == "" {
			continue
		}
		key := strings.ToLower(h)
		if seen[key] {
			warnings = append(warnings, fmt.Sprintf("duplicate column header: %q", h))
		}
		seen[key] = true
	}
	return headers, warnings
}

// isBlankRow reports whether every cell in a row is empty/whitespace.
func isBlankRow(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

// parseExcel parses the first worksheet in an XLSX file and retains every column
// as a contact field.
func (fs *FileService) parseExcel(filePath string) (*models.ParseResult, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("excel file contains no sheets")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel sheet: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("excel sheet is empty")
	}

	if len(rows[0]) > MaxImportColumns {
		return nil, fmt.Errorf("too many columns (%d); limit is %d", len(rows[0]), MaxImportColumns)
	}

	normalizedHeaders, headerWarnings := normalizeHeaders(rows[0])

	// Find column indices (case-insensitive)
	colIndices := fs.findColumnIndices(rows[0])
	if colIndices["email"] == -1 {
		return nil, fmt.Errorf("required column 'Email' not found in Excel file")
	}

	result := &models.ParseResult{
		Contacts: []models.Contact{},
		Errors:   []string{},
		Warnings: append([]string{}, headerWarnings...),
		Headers:  normalizedHeaders,
	}

	// Skip header row
	for i := 1; i < len(rows); i++ {
		if len(result.Contacts) >= MaxImportRows {
			result.Warnings = append(result.Warnings, fmt.Sprintf("import truncated at %d rows", MaxImportRows))
			break
		}
		if isBlankRow(rows[i]) {
			continue
		}
		contact, err := fs.extractContact(rows[i], colIndices, normalizedHeaders, i+1)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
		} else if contact.Email != "" {
			result.Contacts = append(result.Contacts, contact)
		}
	}

	result.Total = len(result.Contacts)
	return result, nil
}

// findColumnIndices scans the header row and finds indices for known columns.
// Matching is case-insensitive and supports common column name variations:
//   - FirstName, First Name, first_name -> "firstname"
//   - LastName, Last Name, last_name -> "lastname"
//   - Email, Email Address, emailaddress -> "email"
//
// Returns a map with keys "firstname", "lastname", "email" and their column indices.
// Index is -1 if column not found.
func (fs *FileService) findColumnIndices(header []string) map[string]int {
	indices := map[string]int{
		"firstname": -1,
		"lastname":  -1,
		"email":     -1,
	}

	for i, col := range header {
		normalized := strings.ToLower(strings.TrimSpace(col))
		for _, sep := range []string{" ", "_", "-", "."} {
			normalized = strings.ReplaceAll(normalized, sep, "")
		}

		// Handle various column name formats.
		switch normalized {
		case "firstname", "fname", "givenname":
			indices["firstname"] = i
		case "lastname", "lname", "surname":
			indices["lastname"] = i
		case "email", "emailaddress", "mail":
			indices["email"] = i
		}
	}

	return indices
}

// extractContact creates a Contact from one imported row, including all mapped
// source columns in CustomFields, and validates the required email address.
//
// Parameters:
//   - row: Array of cell values from the current row
//   - colIndices: Map of field names to column indices
//   - headers: All column headers for custom field names
//   - rowNum: Row number for error reporting (1-based)
//
// Returns:
//   - models.Contact: Extracted contact data with custom fields
//   - error: Validation error if email is missing or invalid
func (fs *FileService) extractContact(row []string, colIndices map[string]int, headers []string, rowNum int) (models.Contact, error) {
	contact := models.Contact{
		ID:           uuid.NewString(),
		CustomFields: make(map[string]string),
	}

	// Helper to safely get column value
	getValue := func(idx int) string {
		if idx >= 0 && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	// Extract standard fields
	contact.FirstName = getValue(colIndices["firstname"])
	contact.LastName = getValue(colIndices["lastname"])
	contact.Email = getValue(colIndices["email"])

	// Extract all columns as custom fields (for merge field support)
	for i, header := range headers {
		if header == "" {
			continue
		}
		value := getValue(i)
		// Store in CustomFields using the original header name (normalized to lowercase for lookup)
		normalizedKey := strings.ToLower(strings.TrimSpace(header))
		contact.CustomFields[normalizedKey] = value
	}

	// Validate email using net/mail (see backend/email) instead of a bare
	// "@"/"." substring check.
	if contact.Email == "" {
		return contact, fmt.Errorf("row %d: missing email address", rowNum)
	}
	if !email.IsValid(contact.Email) {
		return contact, fmt.Errorf("row %d: invalid email format: %s", rowNum, contact.Email)
	}

	return contact, nil
}

// detectDuplicates checks for duplicate email addresses in the parsed contacts.
// Adds duplicate emails to the Duplicates list and creates warnings.
// Does NOT remove duplicates - leaves that decision to the user.
//
// Parameters:
//   - result: ParseResult to check and update with duplicate info
func (fs *FileService) detectDuplicates(result *models.ParseResult) {
	emailCounts := make(map[string]int)

	// Count occurrences of each email
	for _, contact := range result.Contacts {
		email := strings.ToLower(contact.Email)
		emailCounts[email]++
	}

	// Find duplicates
	seen := make(map[string]bool)
	for _, contact := range result.Contacts {
		email := strings.ToLower(contact.Email)
		if emailCounts[email] > 1 && !seen[email] {
			result.Duplicates = append(result.Duplicates, contact.Email)
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Duplicate email found: %s (appears %d times)", contact.Email, emailCounts[email]))
			seen[email] = true
		}
	}
}

// ValidateFile performs pre-parse validation on a file path.
// Checks that the file exists, is not a directory, and has a supported extension.
//
// Parameters:
//   - filePath: Absolute path to the file to validate
//
// Returns:
//   - error: Non-nil if validation fails
func (fs *FileService) ValidateFile(filePath string) error {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}
	if err != nil {
		return fmt.Errorf("failed to access file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}
	if info.Size() > MaxImportFileBytes {
		return fmt.Errorf("file is too large (%d bytes); limit is %d bytes", info.Size(), MaxImportFileBytes)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".csv" && ext != ".xlsx" {
		return fmt.Errorf("unsupported file format: %s. Supported: .csv, .xlsx", ext)
	}

	return nil
}
