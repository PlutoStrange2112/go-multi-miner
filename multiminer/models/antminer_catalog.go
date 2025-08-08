package models

import "strings"

// AntminerModel carries static info we might use for UX and heuristics.
type AntminerModel struct {
	Name       string   // e.g., "S19j Pro"
	Family     string   // e.g., "S19", "S21"
	Cooling    string   // air|hydro|immersion
	Algorithm  string   // usually SHA-256, but include Scrypt/Equihash/Kadena
}

var antminers = []AntminerModel{
	{Name:"S7", Family:"S7", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S9", Family:"S9", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S9k", Family:"S9", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S9 SE", Family:"S9", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S9 Hydro", Family:"S9", Cooling:"hydro", Algorithm:"SHA-256"},
	{Name:"S17", Family:"S17", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S17 Pro", Family:"S17", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S19", Family:"S19", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S19 Pro", Family:"S19", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S19j Pro", Family:"S19", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S19 XP", Family:"S19", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S19 Hydro", Family:"S19", Cooling:"hydro", Algorithm:"SHA-256"},
	{Name:"S19k Pro", Family:"S19", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S19 Pro+ Hyd", Family:"S19", Cooling:"hydro", Algorithm:"SHA-256"},
	{Name:"S21", Family:"S21", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S21 Pro", Family:"S21", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S21+", Family:"S21", Cooling:"air", Algorithm:"SHA-256"},
	{Name:"S21 Hydro", Family:"S21", Cooling:"hydro", Algorithm:"SHA-256"},
	// Non-BTC popular models
	{Name:"L7", Family:"L", Cooling:"air", Algorithm:"Scrypt"},
	{Name:"L9", Family:"L", Cooling:"air", Algorithm:"Scrypt"},
	{Name:"Z15 Pro", Family:"Z", Cooling:"air", Algorithm:"Equihash"},
	{Name:"KA3", Family:"KA", Cooling:"air", Algorithm:"Kadena"},
}

// MatchAntminer scans provided descriptor text (from Version fields) and returns a best-effort model name.
func MatchAntminer(descriptor string) (AntminerModel, bool) {
	s := strings.ToLower(descriptor)
	// Prefer longest/most specific names first
	bestIdx := -1
	bestLen := 0
	for i, m := range antminers {
		if m.Name == "" { continue }
		if idx := strings.Index(s, strings.ToLower(m.Name)); idx >= 0 {
			if l := len(m.Name); l > bestLen { bestLen, bestIdx = l, i }
		}
	}
	if bestIdx >= 0 { return antminers[bestIdx], true }
	// Fallback by family token like "s19" if exact name not present
	families := []string{"s21","s19","s17","s9","l7","l9","z15","ka3"}
	for _, fam := range families {
		if strings.Contains(s, fam) {
			famUp := strings.ToUpper(fam)
			return AntminerModel{Name: famUp, Family: famUp}, true
		}
	}
	return AntminerModel{}, false
}
