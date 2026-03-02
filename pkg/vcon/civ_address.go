package vcon

// CivicAddress contains civic address information for a party's location
// based on RFC 5139 and PIDF-LO format.
type CivicAddress struct {
	// Country code (ISO 3166-1 alpha-2)
	Country string `json:"country,omitempty"`
	// Administrative area 1 (e.g. state or province)
	A1 string `json:"a1,omitempty"`
	// Administrative area 2 (e.g. county or municipality)
	A2 string `json:"a2,omitempty"`
	// Administrative area 3 (e.g. city or town)
	A3 string `json:"a3,omitempty"`
	// Administrative area 4 (e.g. neighborhood or district)
	A4 string `json:"a4,omitempty"`
	// Administrative area 5 (e.g. postal code)
	A5 string `json:"a5,omitempty"`
	// Administrative area 6 (e.g. building or floor)
	A6 string `json:"a6,omitempty"`
	// Premier (e.g. department or suite number)
	PRD string `json:"prd,omitempty"`
	// Post office box identifier
	POD string `json:"pod,omitempty"`
	// Street name
	STS string `json:"sts,omitempty"`
	// House number
	HNO string `json:"hno,omitempty"`
	// House name
	HNS string `json:"hns,omitempty"`
	// Landmark name
	LMK string `json:"lmk,omitempty"`
	// Location name
	LOC string `json:"loc,omitempty"`
	// Floor
	FLR string `json:"flr,omitempty"`
	// Name of the location
	NAM string `json:"nam,omitempty"`
	// Postal code
	PC string `json:"pc,omitempty"`
}

// NewCivicAddress creates a new CivicAddress instance with all fields optional.
func NewCivicAddress() *CivicAddress {
	return &CivicAddress{}
}

type civicField struct {
	key string
	val *string
}

func (c *CivicAddress) fields() []civicField {
	return []civicField{
		{"country", &c.Country},
		{"a1", &c.A1},
		{"a2", &c.A2},
		{"a3", &c.A3},
		{"a4", &c.A4},
		{"a5", &c.A5},
		{"a6", &c.A6},
		{"prd", &c.PRD},
		{"pod", &c.POD},
		{"sts", &c.STS},
		{"hno", &c.HNO},
		{"hns", &c.HNS},
		{"lmk", &c.LMK},
		{"loc", &c.LOC},
		{"flr", &c.FLR},
		{"nam", &c.NAM},
		{"pc", &c.PC},
	}
}

// ToMap converts the CivicAddress object to a map, excluding empty fields.
func (c *CivicAddress) ToMap() map[string]string {
	result := make(map[string]string)
	for _, f := range c.fields() {
		if *f.val != "" {
			result[f.key] = *f.val
		}
	}
	return result
}

// SetFromMap sets CivicAddress fields from a map of strings.
func (c *CivicAddress) SetFromMap(data map[string]string) {
	for _, f := range c.fields() {
		if v, ok := data[f.key]; ok {
			*f.val = v
		}
	}
}
