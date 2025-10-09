// Package codelists provides human-readable descriptions for standard code lists
// used in electronic invoicing (UNTDID, UNECE, etc.).
//
// The code lists are generated from official sources using gencodelists.
package codelists

//go:generate go run ../../../cmd/gencodelists --output generated.go --package codelists

// DocumentType returns the human-readable description for a UNTDID 1001 document type code.
// Returns "Unknown" if the code is not found.
func DocumentType(code string) string {
	if name, ok := documentTypes[code]; ok {
		return name
	}
	return "Unknown"
}

// UnitCode returns the human-readable description for a UNECE Rec 20 unit code.
// Returns the original code if not found.
func UnitCode(code string) string {
	if name, ok := unitCodes[code]; ok {
		return name
	}
	return code
}
