package tax

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
)

// GSTRateInfo represents a single GST rate entry from the Excel file
type GSTRateInfo struct {
	Schedule     string  `json:"schedule"`
	SerialNo     string  `json:"serial_no"`
	HSNCode      string  `json:"hsn_code"`
	Description  string  `json:"description"`
	CGSTRate     float64 `json:"cgst_rate"`
	SGSTRate     float64 `json:"sgst_rate"`
	IGSTRate     float64 `json:"igst_rate"`
	TotalGSTRate float64 `json:"total_gst_rate"`
	Cess         string  `json:"cess"`
}

var (
	gstRates   []GSTRateInfo
	loadOnce   sync.Once
	loadError  error
	excelPath  = "./tax.xlsx"
	sheetName  = "Table 1"
)

// SetExcelPath allows changing the Excel file path before initialization
func SetExcelPath(path string) {
	excelPath = path
}

// Initialize loads GST data from Excel file - must be called before any lookup operations
func Initialize() error {
	loadOnce.Do(func() {
		loadError = loadGSTDataFromExcel()
		if loadError == nil {
			log.Printf("GST database loaded successfully with %d entries", len(gstRates))
		}
	})
	return loadError
}

// loadGSTDataFromExcel reads and parses the GST Excel file
func loadGSTDataFromExcel() error {
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		return fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to read sheet %s: %w", sheetName, err)
	}

	if len(rows) < 2 {
		return fmt.Errorf("Excel file has insufficient data")
	}

	// Parse header to identify column positions
	header := rows[0]
	colMap := make(map[string]int)
	
	for i, col := range header {
		cleanCol := strings.ToLower(strings.TrimSpace(col))
		switch {
		case strings.Contains(cleanCol, "schedule"):
			colMap["schedule"] = i
		case strings.Contains(cleanCol, "s. no"):
			colMap["serial"] = i
		case strings.Contains(cleanCol, "chapter") || strings.Contains(cleanCol, "heading"):
			colMap["hsn"] = i
		case strings.Contains(cleanCol, "description"):
			colMap["description"] = i
		case strings.Contains(cleanCol, "cgst"):
			colMap["cgst"] = i
		case strings.Contains(cleanCol, "sgst") || strings.Contains(cleanCol, "utgst"):
			colMap["sgst"] = i
		case strings.Contains(cleanCol, "igst"):
			colMap["igst"] = i
		case strings.Contains(cleanCol, "compensation"):
			colMap["cess"] = i
		}
	}

	// Clear existing data and parse all rows
	gstRates = make([]GSTRateInfo, 0)
	
	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		
		// Helper function to safely get cell value
		getCell := func(colName string) string {
			if idx, exists := colMap[colName]; exists && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		schedule := getCell("schedule")
		serialNo := getCell("serial")
		hsn := getCell("hsn")
		description := getCell("description")
		cgstStr := getCell("cgst")
		sgstStr := getCell("sgst")
		igstStr := getCell("igst")
		cess := getCell("cess")

		// Skip rows with no meaningful data
		if hsn == "" && description == "" && cgstStr == "" && igstStr == "" {
			continue
		}

		// Skip omitted entries
		if strings.Contains(strings.ToLower(description), "omitted") {
			continue
		}

		// Parse GST rates
		cgstRate := parseRate(cgstStr)
		sgstRate := parseRate(sgstStr)
		igstRate := parseRate(igstStr)

		// Calculate total GST rate (IGST or CGST+SGST)
		totalRate := igstRate
		if totalRate == 0 && (cgstRate > 0 || sgstRate > 0) {
			totalRate = cgstRate + sgstRate
		}

		gstInfo := GSTRateInfo{
			Schedule:     schedule,
			SerialNo:     serialNo,
			HSNCode:      hsn,
			Description:  description,
			CGSTRate:     cgstRate,
			SGSTRate:     sgstRate,
			IGSTRate:     igstRate,
			TotalGSTRate: totalRate,
			Cess:         cess,
		}

		gstRates = append(gstRates, gstInfo)
	}

	log.Printf("Successfully loaded %d GST rate entries from Excel", len(gstRates))
	return nil
}

// parseRate converts rate string to float64 percentage
func parseRate(rateStr string) float64 {
	if rateStr == "" {
		return 0.0
	}
	
	// Clean the rate string
	rateStr = strings.TrimSpace(rateStr)
	rateStr = strings.ReplaceAll(rateStr, "%", "")
	
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		return 0.0
	}
	
	// Convert to percentage if it's in decimal form (like 0.025 -> 2.5%)
	if rate < 1.0 && rate > 0 {
		rate = rate * 100
	}
	
	return rate
}

// GetGSTRateByHSN looks up GST rate by HSN code
func GetGSTRateByHSN(hsnCode string) *GSTRateInfo {
	if len(gstRates) == 0 {
		return nil
	}

	hsnCode = strings.TrimSpace(hsnCode)
	if hsnCode == "" {
		return nil
	}

	for _, entry := range gstRates {
		if matchesHSN(entry.HSNCode, hsnCode) {
			entryCopy := entry
			return &entryCopy
		}
	}
	return nil
}

// GetGSTRateByCategory looks up GST rate by product category/description
func GetGSTRateByCategory(category string) *GSTRateInfo {
	if len(gstRates) == 0 {
		return nil
	}

	category = strings.ToLower(strings.TrimSpace(category))
	if category == "" {
		return nil
	}

	for _, entry := range gstRates {
		desc := strings.ToLower(entry.Description)
		if strings.Contains(desc, category) {
			entryCopy := entry
			return &entryCopy
		}
	}
	return nil
}

// GetGSTRateByDescription looks up GST rate by description keywords
func GetGSTRateByDescription(description string) *GSTRateInfo {
	if len(gstRates) == 0 {
		return nil
	}

	description = strings.ToLower(strings.TrimSpace(description))
	if description == "" {
		return nil
	}

	// Split description into keywords and find best match
	keywords := strings.Fields(description)
	
	bestMatch := (*GSTRateInfo)(nil)
	maxMatches := 0

	for _, entry := range gstRates {
		entryDesc := strings.ToLower(entry.Description)
		matches := 0
		
		for _, keyword := range keywords {
			if len(keyword) > 2 && strings.Contains(entryDesc, keyword) {
				matches++
			}
		}
		
		if matches > maxMatches {
			maxMatches = matches
			entryCopy := entry
			bestMatch = &entryCopy
		}
	}

	return bestMatch
}

// GetGSTRate is the main function to get GST rate for a product
func GetGSTRate(hsnCode, category, description string) *GSTRateInfo {
	// Try HSN code first (most accurate)
	if hsnCode != "" {
		if rate := GetGSTRateByHSN(hsnCode); rate != nil {
			return rate
		}
	}

	// Try category
	if category != "" {
		if rate := GetGSTRateByCategory(category); rate != nil {
			return rate
		}
	}

	// Try description
	if description != "" {
		if rate := GetGSTRateByDescription(description); rate != nil {
			return rate
		}
	}

	// Return default rate if nothing found (18% GST)
	return &GSTRateInfo{
		Schedule:     "Default",
		HSNCode:      "0000",
		Description:  "Default rate",
		CGSTRate:     9.0,
		SGSTRate:     9.0,
		IGSTRate:     18.0,
		TotalGSTRate: 18.0,
	}
}

// GetGSTRateForProduct is a convenience function for use in GORM BeforeSave
func GetGSTRateForProduct(gstCategory, hsnCode, productName, countryCode string) float64 {
	gstInfo := GetGSTRate(hsnCode, gstCategory, productName)
	if gstInfo == nil {
		return 18.0 // Default GST rate
	}

	// For India, return IGST rate
	if countryCode == "IN" {
		return gstInfo.IGSTRate
	}

	return 0.0 // No GST for non-Indian products
}

// matchesHSN checks if the HSN from database matches the provided HSN
func matchesHSN(dbHSN, searchHSN string) bool {
	if dbHSN == "" || searchHSN == "" {
		return false
	}

	// Clean both HSN codes
	dbHSN = strings.ReplaceAll(strings.TrimSpace(dbHSN), " ", "")
	searchHSN = strings.ReplaceAll(strings.TrimSpace(searchHSN), " ", "")

	// Handle comma-separated HSN codes in database
	hsnCodes := strings.Split(dbHSN, ",")
	for _, code := range hsnCodes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}

		// Check for exact match or partial match
		if code == searchHSN || strings.HasPrefix(searchHSN, code) || strings.HasPrefix(code, searchHSN) {
			return true
		}
	}

	return false
}

// GetAllGSTRates returns all GST rates (for admin/debugging purposes)
func GetAllGSTRates() []GSTRateInfo {
	return gstRates
}

// GetGSTRateDetails returns detailed GST information for a product
func GetGSTRateDetails(hsnCode, category, description string) map[string]interface{} {
	gstInfo := GetGSTRate(hsnCode, category, description)
	if gstInfo == nil {
		return map[string]interface{}{
			"found": false,
			"error": "No GST rate found",
		}
	}

	return map[string]interface{}{
		"found":           true,
		"schedule":        gstInfo.Schedule,
		"hsn_code":        gstInfo.HSNCode,
		"description":     gstInfo.Description,
		"cgst_rate":       gstInfo.CGSTRate,
		"sgst_rate":       gstInfo.SGSTRate,
		"igst_rate":       gstInfo.IGSTRate,
		"total_gst_rate":  gstInfo.TotalGSTRate,
		"cess":            gstInfo.Cess,
	}
}

// SearchGSTRates searches for GST rates by keyword
func SearchGSTRates(keyword string) []GSTRateInfo {
	if len(gstRates) == 0 {
		return nil
	}

	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return nil
	}

	var results []GSTRateInfo
	for _, entry := range gstRates {
		if strings.Contains(strings.ToLower(entry.Description), keyword) ||
			strings.Contains(strings.ToLower(entry.HSNCode), keyword) {
			results = append(results, entry)
		}
	}

	return results
}
