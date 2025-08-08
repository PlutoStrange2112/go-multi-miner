package models

import "strings"

type WhatsminerModel struct {
	Name      string
	Cooling   string
	Notes     string
}

var whatsminers = []WhatsminerModel{
	{Name:"M1", Cooling:"air"},
	{Name:"M3", Cooling:"air"},
	{Name:"M10", Cooling:"air"},
	{Name:"M20S", Cooling:"air"},
	{Name:"M21S", Cooling:"air"},
	{Name:"M30S", Cooling:"air"},
	{Name:"M30S+", Cooling:"air"},
	{Name:"M30S++", Cooling:"air"},
	{Name:"M31S", Cooling:"air"},
	{Name:"M31S+", Cooling:"air"},
	{Name:"M32", Cooling:"air"},
	{Name:"M50", Cooling:"air"},
	{Name:"M50S", Cooling:"air"},
	{Name:"M50S+", Cooling:"air"},
	{Name:"M50S++", Cooling:"air"},
	{Name:"M53", Cooling:"hydro"},
	{Name:"M53S++", Cooling:"hydro"},
	{Name:"M56S++", Cooling:"immersion"},
	{Name:"M60", Cooling:"air"},
	{Name:"M60S", Cooling:"air"},
	{Name:"M63", Cooling:"hydro"},
	{Name:"M63S", Cooling:"hydro"},
	{Name:"M66", Cooling:"immersion"},
	{Name:"M66S", Cooling:"immersion"},
	{Name:"M70", Cooling:"air"},
}

func MatchWhatsminer(desc string) (WhatsminerModel, bool) {
	s := strings.ToLower(desc)
	bestIdx := -1
	bestLen := 0
	for i, m := range whatsminers {
		if idx := strings.Index(s, strings.ToLower(m.Name)); idx >= 0 {
			if l := len(m.Name); l > bestLen { bestLen, bestIdx = l, i }
		}
	}
	if bestIdx >= 0 { return whatsminers[bestIdx], true }
	return WhatsminerModel{}, false
}
