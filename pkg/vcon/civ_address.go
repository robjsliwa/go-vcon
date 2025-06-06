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

// ToMap converts the CivicAddress object to a map, excluding empty fields.
// This is similar to the Python to_dict() method.
func (c *CivicAddress) ToMap() map[string]string {
	result := make(map[string]string)
	
	// Only include non-empty fields in the map
	if c.Country != "" {
		result["country"] = c.Country
	}
	if c.A1 != "" {
		result["a1"] = c.A1
	}
	if c.A2 != "" {
		result["a2"] = c.A2
	}
	if c.A3 != "" {
		result["a3"] = c.A3
	}
	if c.A4 != "" {
		result["a4"] = c.A4
	}
	if c.A5 != "" {
		result["a5"] = c.A5
	}
	if c.A6 != "" {
		result["a6"] = c.A6
	}
	if c.PRD != "" {
		result["prd"] = c.PRD
	}
	if c.POD != "" {
		result["pod"] = c.POD
	}
	if c.STS != "" {
		result["sts"] = c.STS
	}
	if c.HNO != "" {
		result["hno"] = c.HNO
	}
	if c.HNS != "" {
		result["hns"] = c.HNS
	}
	if c.LMK != "" {
		result["lmk"] = c.LMK
	}
	if c.LOC != "" {
		result["loc"] = c.LOC
	}
	if c.FLR != "" {
		result["flr"] = c.FLR
	}
	if c.NAM != "" {
		result["nam"] = c.NAM
	}
	if c.PC != "" {
		result["pc"] = c.PC
	}
	
	return result
}

// SetFromMap sets CivicAddress fields from a map of strings.
func (c *CivicAddress) SetFromMap(data map[string]string) {
	if country, ok := data["country"]; ok {
		c.Country = country
	}
	if a1, ok := data["a1"]; ok {
		c.A1 = a1
	}
	if a2, ok := data["a2"]; ok {
		c.A2 = a2
	}
	if a3, ok := data["a3"]; ok {
		c.A3 = a3
	}
	if a4, ok := data["a4"]; ok {
		c.A4 = a4
	}
	if a5, ok := data["a5"]; ok {
		c.A5 = a5
	}
	if a6, ok := data["a6"]; ok {
		c.A6 = a6
	}
	if prd, ok := data["prd"]; ok {
		c.PRD = prd
	}
	if pod, ok := data["pod"]; ok {
		c.POD = pod
	}
	if sts, ok := data["sts"]; ok {
		c.STS = sts
	}
	if hno, ok := data["hno"]; ok {
		c.HNO = hno
	}
	if hns, ok := data["hns"]; ok {
		c.HNS = hns
	}
	if lmk, ok := data["lmk"]; ok {
		c.LMK = lmk
	}
	if loc, ok := data["loc"]; ok {
		c.LOC = loc
	}
	if flr, ok := data["flr"]; ok {
		c.FLR = flr
	}
	if nam, ok := data["nam"]; ok {
		c.NAM = nam
	}
	if pc, ok := data["pc"]; ok {
		c.PC = pc
	}
}
