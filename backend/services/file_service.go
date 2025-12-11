/*
File Service - Contact File Import

This service handles importing contact lists from CSV and Excel files.
It provides a unified interface for parsing both file formats and
extracting contact information.

Supported Formats:
  - CSV (.csv) - Comma-separated values with header row
  - Excel (.xlsx) - Microsoft Excel 2007+ format (first sheet only)

Phase 1 Updates (v1.2):
  - Parses ALL columns from the file, not just FirstName/LastName/Email
  - Returns all column headers for dynamic merge field generation
  - Detects and reports duplicate email addresses
  - Stores custom fields in Contact.CustomFields map

Column Detection:
  - Case-insensitive column name matching
  - Supports variations: "FirstName", "First Name", "first_name"
  - Email column is required; all other columns are optional

Error Handling:
  - Returns parse errors for invalid rows without failing entire import
  - Reports duplicate emails as warnings (still includes contacts)
  - Validates email format with basic @ and . check
*/
package services

import (
	"MailMergeApp/backend/models"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// FileService handles importing contacts from CSV and Excel files.
// Stateless service - each parse operation is independent.
type FileService struct{}

// NewFileService creates a new FileService instance.
func NewFileService() *FileService {
	return &FileService{}
}

// ParseContactFile parses a contact file and extracts contact information.
// Automatically detects file format based on extension.
// Phase 1: Now returns all column headers and detects duplicates.
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

// parseCSV parses a CSV file and extracts contacts.
// Expects a header row with column names.
// Phase 1: Now extracts all columns as custom fields.
func (fs *FileService) parseCSV(filePath string) (*models.ParseResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header row
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Normalize and store all headers
	normalizedHeaders := make([]string, len(header))
	for i, h := range header {
		normalizedHeaders[i] = strings.TrimSpace(h)
	}

	// Find column indices (case-insensitive)
	colIndices := fs.findColumnIndices(header)
	if colIndices["email"] == -1 {
		return nil, fmt.Errorf("required column 'Email' not found in CSV")
	}

	result := &models.ParseResult{
		Contacts: []models.Contact{},
		Errors:   []string{},
		Warnings: []string{},
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

// parseExcel parses an Excel file (.xlsx) and extracts contacts.
// Only reads from the first sheet in the workbook.
// Phase 1: Now extracts all columns as custom fields.
func (fs *FileService) parseExcel(filePath string) (*models.ParseResult, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel file contains no sheets")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel sheet: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("Excel sheet is empty")
	}

	// Normalize and store all headers
	normalizedHeaders := make([]string, len(rows[0]))
	for i, h := range rows[0] {
		normalizedHeaders[i] = strings.TrimSpace(h)
	}

	// Find column indices (case-insensitive)
	colIndices := fs.findColumnIndices(rows[0])
	if colIndices["email"] == -1 {
		return nil, fmt.Errorf("required column 'Email' not found in Excel file")
	}

	result := &models.ParseResult{
		Contacts: []models.Contact{},
		Errors:   []string{},
		Warnings: []string{},
		Headers:  normalizedHeaders,
	}

	// Skip header row
	for i := 1; i < len(rows); i++ {
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
		normalized = strings.ReplaceAll(normalized, " ", "")
		normalized = strings.ReplaceAll(normalized, "_", "")

		// Handle various column name formats
		switch {
		case normalized == "firstname" || normalized == "fname":
			indices["firstname"] = i
		case normalized == "lastname" || normalized == "lname":
			indices["lastname"] = i
		case normalized == "email" || normalized == "emailaddress" || normalized == "mail":
			indices["email"] = i
		}
	}

	return indices
}

// extractContact creates a Contact from a data row using the provided column indices.
// Phase 1: Now populates CustomFields with all columns beyond FirstName/LastName/Email.
// Validates that email is present and has a valid format.
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

	// Validate email
	if contact.Email == "" {
		return contact, fmt.Errorf("row %d: missing email address", rowNum)
	}

	// Basic email validation
	if !strings.Contains(contact.Email, "@") || !strings.Contains(contact.Email, ".") {
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

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".csv" && ext != ".xlsx" {
		return fmt.Errorf("unsupported file format: %s. Supported: .csv, .xlsx", ext)
	}

	return nil
}
