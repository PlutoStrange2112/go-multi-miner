package cgminer

import (
	"fmt"
)

type Command struct {
	Command   string `json:"command"`
	Parameter string `json:"parameter,omitempty"`
}

func NewCommandWithoutParameter(name string) Command {
	return Command{Command: name}
}

func NewCommand(name, parameter string) Command {
	return Command{
		Command:   name,
		Parameter: parameter,
	}
}

// GenericResponse - default struct for all responses
type GenericResponse struct {
	ID     int      `json:"id"`
	Status []Status `json:"STATUS"`
}

// HasError implements AbstractResponse interface
func (r GenericResponse) HasError() error {
	for _, status := range r.Status {
		switch status.Status {
		case "E":
			return fmt.Errorf("API returned error: Code: %d, Msg: '%s', Description: '%s'", status.Code, status.Msg, status.Description)
		case "F":
			return fmt.Errorf("API returned FATAL error: Code: %d, Msg: '%s', Description: '%s'", status.Code, status.Msg, status.Description)
		}
	}
	return nil
}

// VersionResponse - returned by "version" command
type VersionResponse struct {
	GenericResponse
	Version []Version `json:"VERSION"`
}

// Status - status of request
type Status struct {
	Status      string `json:"STATUS"`
	When        int
	Code        int
	Msg         string
	Description string
}

// Version - version of miner software and hw model
type Version struct {
	BMMiner     string
	API         string
	Miner       string
	CompileTime string
	Type        string
}

// GenericStats - generic antminer stats struct
type GenericStats struct {
	BMMiner               string
	CGMiner               string `json:"CGMiner,omitempty"`
	Miner                 string
	CompileTime           string
	Type                  string
	MinerID               string `json:"miner_id"`
	MinerVersion          string `json:"miner_version"`
	MinerCount            int16  `json:"miner_count"`
	Elapsed               int64  `json:"Elapsed"`
	Wait                  float64
	DeviceHardwarePercent float64 `json:"Device Hardware%"`
	Stats                 int     `json:"STATS"`
	Max                   float64
	NotMatchingWork       int `json:"no_matching_work"`
	ID                    string
	Calls                 int
	Min                   float64
	TotalAcn              int16   `json:"total_acn"`
	TotalRate             float32 `json:"total_rate"`
	TotalRateIdeal        float32 `json:"total_rateideal"`
	TotalFrequencyAvg     float32 `json:"total_freqavg"`
	Frequency             Number  `json:"frequency"`
	FrequencyAvg1         float32 `json:"freq_avg1"`
	FrequencyAvg2         float32 `json:"freq_avg2"`
	FrequencyAvg3         float32 `json:"freq_avg3"`
	FrequencyAvg4         float32 `json:"freq_avg4"`
	FrequencyAvg5         float32 `json:"freq_avg5"`
	FrequencyAvg6         float32 `json:"freq_avg6"`
	FrequencyAvg7         float32 `json:"freq_avg7"`
	FrequencyAvg8         float32 `json:"freq_avg8"`
	FrequencyAvg9         float32 `json:"freq_avg9"`
	FrequencyAvg10        float32 `json:"freq_avg10"`
	FrequencyAvg11        float32 `json:"freq_avg11"`
	FrequencyAvg12        float32 `json:"freq_avg12"`
	FrequencyAvg13        float32 `json:"freq_avg13"`
	FrequencyAvg14        float32 `json:"freq_avg14"`
	FrequencyAvg15        float32 `json:"freq_avg15"`
	FrequencyAvg16        float32 `json:"freq_avg16"`
	FanNum                int16   `json:"fan_num"`
	Fan1                  int16   `json:"fan1"`
	Fan2                  int16   `json:"fan2"`
	Fan3                  int16   `json:"fan3"`
	Fan4                  int16   `json:"fan4"`
	Fan5                  int16   `json:"fan5"`
	Fan6                  int16   `json:"fan6"`
	Fan7                  int16   `json:"fan7"`
	Fan8                  int16   `json:"fan8"`
	TempMax               int16   `json:"temp_max"`
	TempNum               int16   `json:"temp_num"`
	Temp1                 int16   `json:"temp1"`
	Temp2                 int16   `json:"temp2"`
	Temp3                 int16   `json:"temp3"`
	Temp4                 int16   `json:"temp4"`
	Temp5                 int16   `json:"temp5"`
	Temp6                 int16   `json:"temp6"`
	Temp7                 int16   `json:"temp7"`
	Temp8                 int16   `json:"temp8"`
	Temp9                 int16   `json:"temp9"`
	Temp10                int16   `json:"temp10"`
	Temp11                int16   `json:"temp11"`
	Temp12                int16   `json:"temp12"`
	Temp13                int16   `json:"temp13"`
	Temp14                int16   `json:"temp14"`
	Temp15                int16   `json:"temp15"`
	Temp16                int16   `json:"temp16"`
	Temp2_1               int16   `json:"temp2_1"`
	Temp2_2               int16   `json:"temp2_2"`
	Temp2_3               int16   `json:"temp2_3"`
	Temp2_4               int16   `json:"temp2_4"`
	Temp2_5               int16   `json:"temp2_5"`
	Temp2_6               int16   `json:"temp2_6"`
	Temp2_7               int16   `json:"temp2_7"`
	Temp2_8               int16   `json:"temp2_8"`
	Temp2_9               int16   `json:"temp2_9"`
	Temp2_10              int16   `json:"temp2_10"`
	Temp2_11              int16   `json:"temp2_11"`
	Temp2_12              int16   `json:"temp2_12"`
	Temp2_13              int16   `json:"temp2_13"`
	Temp2_14              int16   `json:"temp2_14"`
	Temp2_15              int16   `json:"temp2_15"`
	Temp2_16              int16   `json:"temp2_16"`
	Temp3_1               int16   `json:"temp3_1"`
	Temp3_2               int16   `json:"temp3_2"`
	Temp3_3               int16   `json:"temp3_3"`
	Temp3_4               int16   `json:"temp3_4"`
	Temp3_5               int16   `json:"temp3_5"`
	Temp3_6               int16   `json:"temp3_6"`
	Temp3_7               int16   `json:"temp3_7"`
	Temp3_8               int16   `json:"temp3_8"`
	Temp3_9               int16   `json:"temp3_9"`
	Temp3_10              int16   `json:"temp3_10"`
	Temp3_11              int16   `json:"temp3_11"`
	Temp3_12              int16   `json:"temp3_12"`
	Temp3_13              int16   `json:"temp3_13"`
	Temp3_14              int16   `json:"temp3_14"`
	Temp3_15              int16   `json:"temp3_15"`
	Temp3_16              int16   `json:"temp3_16"`
	// this from L3+
	Temp31  int16 `json:"temp31"`
	Temp32  int16 `json:"temp32"`
	Temp33  int16 `json:"temp33"`
	Temp34  int16 `json:"temp34"`
	Temp4_1 int16 `json:"temp4_1"`
	Temp4_2 int16 `json:"temp4_2"`
	Temp4_3 int16 `json:"temp4_3"`
	Temp4_4 int16 `json:"temp4_4"`
	// interfcace{} - because S7 has valid value, but all others have quoted value
	Ghs5s      Number  `json:"GHS 5s"`
	GhsAverage float64 `json:"GHS av"`
	ChainHW1   int     `json:"chain_hw1"`
	ChainHW2   int     `json:"chain_hw2"`
	ChainHW3   int     `json:"chain_hw3"`
	ChainHW4   int     `json:"chain_hw4"`
	ChainHW5   int     `json:"chain_hw5"`
	ChainHW6   int     `json:"chain_hw6"`
	ChainHW7   int     `json:"chain_hw7"`
	ChainHW8   int     `json:"chain_hw8"`
	ChainHW9   int     `json:"chain_hw9"`
	ChainHW10  int     `json:"chain_hw10"`
	ChainHW11  int     `json:"chain_hw11"`
	ChainHW12  int     `json:"chain_hw12"`
	ChainHW13  int     `json:"chain_hw13"`
	ChainHW14  int     `json:"chain_hw14"`
	ChainHW15  int     `json:"chain_hw15"`
	ChainHW16  int     `json:"chain_hw16"`
	ChainAcs1  string  `json:"chain_acs1"`
	ChainAcs2  string  `json:"chain_acs2"`
	ChainAcs3  string  `json:"chain_acs3"`
	ChainAcs4  string  `json:"chain_acs4"`
	ChainAcs5  string  `json:"chain_acs5"`
	ChainAcs6  string  `json:"chain_acs6"`
	ChainAcs7  string  `json:"chain_acs7"`
	ChainAcs8  string  `json:"chain_acs8"`
	ChainAcs9  string  `json:"chain_acs9"`
	ChainAcs10 string  `json:"chain_acs10"`
	ChainAcs11 string  `json:"chain_acs11"`
	ChainAcs12 string  `json:"chain_acs12"`
	ChainAcs13 string  `json:"chain_acs13"`
	ChainAcs14 string  `json:"chain_acs14"`
	ChainAcs15 string  `json:"chain_acs15"`
	ChainAcs16 string  `json:"chain_acs16"`
	ChainAcn1  int     `json:"chain_acn1"`
	ChainAcn2  int     `json:"chain_acn2"`
	ChainAcn3  int     `json:"chain_acn3"`
	ChainAcn4  int     `json:"chain_acn4"`
	ChainAcn5  int     `json:"chain_acn5"`
	ChainAcn6  int     `json:"chain_acn6"`
	ChainAcn7  int     `json:"chain_acn7"`
	ChainAcn8  int     `json:"chain_acn8"`
	ChainAcn9  int     `json:"chain_acn9"`
	ChainAcn10 int     `json:"chain_acn10"`
	ChainAcn11 int     `json:"chain_acn11"`
	ChainAcn12 int     `json:"chain_acn12"`
	ChainAcn13 int     `json:"chain_acn13"`
	ChainAcn14 int     `json:"chain_acn14"`
	ChainAcn15 int     `json:"chain_acn15"`
	ChainAcn16 int     `json:"chain_acn16"`
	// Number because we cannot parse empty string "" to float
	// S9(), D3() and other interface methods will correct represent this value as float32
	ChainRate1       Number  `json:"chain_rate1"`
	ChainRate2       Number  `json:"chain_rate2"`
	ChainRate3       Number  `json:"chain_rate3"`
	ChainRate4       Number  `json:"chain_rate4"`
	ChainRate5       Number  `json:"chain_rate5"`
	ChainRate6       Number  `json:"chain_rate6"`
	ChainRate7       Number  `json:"chain_rate7"`
	ChainRate8       Number  `json:"chain_rate8"`
	ChainRate9       Number  `json:"chain_rate9"`
	ChainRate10      Number  `json:"chain_rate10"`
	ChainRate11      Number  `json:"chain_rate11"`
	ChainRate12      Number  `json:"chain_rate12"`
	ChainRate13      Number  `json:"chain_rate13"`
	ChainRate14      Number  `json:"chain_rate14"`
	ChainRate15      Number  `json:"chain_rate15"`
	ChainRate16      Number  `json:"chain_rate16"`
	ChainRateIdeal1  float32 `json:"chain_rateideal1"`
	ChainRateIdeal2  float32 `json:"chain_rateideal2"`
	ChainRateIdeal3  float32 `json:"chain_rateideal3"`
	ChainRateIdeal4  float32 `json:"chain_rateideal4"`
	ChainRateIdeal5  float32 `json:"chain_rateideal5"`
	ChainRateIdeal6  float32 `json:"chain_rateideal6"`
	ChainRateIdeal7  float32 `json:"chain_rateideal7"`
	ChainRateIdeal8  float32 `json:"chain_rateideal8"`
	ChainRateIdeal9  float32 `json:"chain_rateideal9"`
	ChainRateIdeal10 float32 `json:"chain_rateideal10"`
	ChainRateIdeal11 float32 `json:"chain_rateideal11"`
	ChainRateIdeal12 float32 `json:"chain_rateideal12"`
	ChainRateIdeal13 float32 `json:"chain_rateideal13"`
	ChainRateIdeal14 float32 `json:"chain_rateideal14"`
	ChainRateIdeal15 float32 `json:"chain_rateideal15"`
	ChainRateIdeal16 float32 `json:"chain_rateideal16"`
	ChainOpenCore1   Number  `json:"chain_opencore_1"`
	ChainOpenCore2   Number  `json:"chain_opencore_2"`
	ChainOpenCore3   Number  `json:"chain_opencore_3"`
	ChainOpenCore4   Number  `json:"chain_opencore_4"`
	ChainOpenCore5   Number  `json:"chain_opencore_5"`
	ChainOpenCore6   Number  `json:"chain_opencore_6"`
	ChainOpenCore7   Number  `json:"chain_opencore_7"`
	ChainOpenCore8   Number  `json:"chain_opencore_8"`
	ChainOpenCore9   Number  `json:"chain_opencore_9"`
	ChainOpenCore10  Number  `json:"chain_opencore_10"`
	ChainOpenCore11  Number  `json:"chain_opencore_11"`
	ChainOpenCore12  Number  `json:"chain_opencore_12"`
	ChainOpenCore13  Number  `json:"chain_opencore_13"`
	ChainOpenCore14  Number  `json:"chain_opencore_14"`
	ChainOpenCore15  Number  `json:"chain_opencore_15"`
	ChainOpenCore16  Number  `json:"chain_opencore_16"`
	ChainOffside1    Number  `json:"chain_offside_1"`
	ChainOffside2    Number  `json:"chain_offside_2"`
	ChainOffside3    Number  `json:"chain_offside_3"`
	ChainOffside4    Number  `json:"chain_offside_4"`
	ChainOffside5    Number  `json:"chain_offside_5"`
	ChainOffside6    Number  `json:"chain_offside_6"`
	ChainOffside7    Number  `json:"chain_offside_7"`
	ChainOffside8    Number  `json:"chain_offside_8"`
	ChainOffside9    Number  `json:"chain_offside_9"`
	ChainOffside10   Number  `json:"chain_offside_10"`
	ChainOffside11   Number  `json:"chain_offside_11"`
	ChainOffside12   Number  `json:"chain_offside_12"`
	ChainOffside13   Number  `json:"chain_offside_13"`
	ChainOffside14   Number  `json:"chain_offside_14"`
	ChainOffside15   Number  `json:"chain_offside_15"`
	ChainOffside16   Number  `json:"chain_offside_16"`
	ChainXtime1      string  `json:"chain_xtime1"`
	ChainXtime2      string  `json:"chain_xtime2"`
	ChainXtime3      string  `json:"chain_xtime3"`
	ChainXtime4      string  `json:"chain_xtime4"`
	ChainXtime5      string  `json:"chain_xtime5"`
	ChainXtime6      string  `json:"chain_xtime6"`
	ChainXtime7      string  `json:"chain_xtime7"`
	ChainXtime8      string  `json:"chain_xtime8"`
	ChainXtime9      string  `json:"chain_xtime9"`
	ChainXtime10     string  `json:"chain_xtime10"`
	ChainXtime11     string  `json:"chain_xtime11"`
	ChainXtime12     string  `json:"chain_xtime12"`
	ChainXtime13     string  `json:"chain_xtime13"`
	ChainXtime14     string  `json:"chain_xtime14"`
	ChainXtime15     string  `json:"chain_xtime15"`
	ChainXtime16     string  `json:"chain_xtime16"`
	// s7
	Baud      int    `json:"baud"`
	AsicCount int    `json:"asic_count"`
	Timeout   int    `json:"timeout"`
	Voltage   Number `json:"voltage"`
	USBPipe   Number `json:"USB Pipe"`
	HWv1      int    `json:"hwv1"`
	HWv2      int    `json:"hwv2"`
	HWv3      int    `json:"hwv3"`
	HWv4      int    `json:"hwv4"`
	TempAvg   int16  `json:"temp_avg"`
}

// StatsS7 - generic antminer stats struct
type StatsS7 struct {
	CGMiner               string `json:"CGMiner,omitempty"`
	Miner                 string
	CompileTime           string
	Type                  string
	Stats                 int `json:"STATS"`
	ID                    string
	Elapsed               int64 `json:"Elapsed"`
	Calls                 int
	Wait                  float64
	Max                   float64
	Min                   float64
	Ghs5s                 Number  `json:"GHS 5s"`
	GhsAverage            float64 `json:"GHS av"`
	Baud                  int     `json:"baud"`
	MinerCount            int16   `json:"miner_count"`
	AsicCount             int     `json:"asic_count"`
	Timeout               int     `json:"timeout"`
	Frequency             Number  `json:"frequency"`
	Voltage               float32 `json:"voltage"`
	FanNum                int16   `json:"fan_num"`
	Fan1                  int16   `json:"fan1"`
	Fan3                  int16   `json:"fan3"`
	TempNum               int16   `json:"temp_num"`
	Temp1                 int16   `json:"temp1"`
	Temp2                 int16   `json:"temp2"`
	Temp3                 int16   `json:"temp3"`
	TempAvg               int16   `json:"temp_avg"`
	TempMax               int16   `json:"temp_max"`
	DeviceHardwarePercent float64 `json:"Device Hardware%"`
	NotMatchingWork       int     `json:"no_matching_work"`
	USBPipe               Number  `json:"USB Pipe"`
	HWv1                  int     `json:"hwv1"`
	HWv2                  int     `json:"hwv2"`
	HWv3                  int     `json:"hwv3"`
	HWv4                  int     `json:"hwv4"`
	ChainAcn1             int     `json:"chain_acn1"`
	ChainAcn2             int     `json:"chain_acn2"`
	ChainAcn3             int     `json:"chain_acn3"`
	ChainAcs1             string  `json:"chain_acs1"`
	ChainAcs2             string  `json:"chain_acs2"`
	ChainAcs3             string  `json:"chain_acs3"`
}

// StatsT9 - generic antminer stats struct
type StatsT9 struct {
	BMMiner               string
	Miner                 string
	CompileTime           string
	Type                  string
	ID                    string
	Stats                 int   `json:"STATS"`
	Elapsed               int64 `json:"Elapsed"`
	Calls                 int
	Wait                  float64
	Max                   float64
	Min                   float64
	Ghs5s                 Number  `json:"GHS 5s"`
	GhsAverage            float64 `json:"GHS av"`
	MinerCount            int16   `json:"miner_count"`
	Frequency             Number  `json:"frequency"`
	FanNum                int16   `json:"fan_num"`
	Fan3                  int16   `json:"fan3"`
	Fan6                  int16   `json:"fan6"`
	TempNum               int16   `json:"temp_num"`
	Temp2                 int16   `json:"temp2"`
	Temp3                 int16   `json:"temp3"`
	Temp4                 int16   `json:"temp4"`
	Temp9                 int16   `json:"temp9"`
	Temp10                int16   `json:"temp10"`
	Temp11                int16   `json:"temp11"`
	Temp12                int16   `json:"temp12"`
	Temp13                int16   `json:"temp13"`
	Temp14                int16   `json:"temp14"`
	Temp2_2               int16   `json:"temp2_2"`
	Temp2_3               int16   `json:"temp2_3"`
	Temp2_4               int16   `json:"temp2_4"`
	Temp2_9               int16   `json:"temp2_9"`
	Temp2_10              int16   `json:"temp2_10"`
	Temp2_11              int16   `json:"temp2_11"`
	Temp2_12              int16   `json:"temp2_12"`
	Temp2_13              int16   `json:"temp2_13"`
	Temp2_14              int16   `json:"temp2_14"`
	TempMax               int16   `json:"temp_max"`
	FrequencyAvg2         float32 `json:"freq_avg2"`
	FrequencyAvg3         float32 `json:"freq_avg3"`
	FrequencyAvg4         float32 `json:"freq_avg4"`
	FrequencyAvg5         float32 `json:"freq_avg5"`
	FrequencyAvg6         float32 `json:"freq_avg6"`
	FrequencyAvg7         float32 `json:"freq_avg7"`
	FrequencyAvg8         float32 `json:"freq_avg8"`
	FrequencyAvg9         float32 `json:"freq_avg9"`
	FrequencyAvg10        float32 `json:"freq_avg10"`
	FrequencyAvg11        float32 `json:"freq_avg11"`
	FrequencyAvg12        float32 `json:"freq_avg12"`
	FrequencyAvg13        float32 `json:"freq_avg13"`
	FrequencyAvg14        float32 `json:"freq_avg14"`
	TotalRateIdeal        float32 `json:"total_rateideal"`
	TotalFrequencyAvg     float32 `json:"total_freqavg"`
	TotalAcn              int16   `json:"total_acn"`
	TotalRate             float32 `json:"total_rate"`
	ChainRateIdeal2       float32 `json:"chain_rateideal2"`
	ChainRateIdeal3       float32 `json:"chain_rateideal3"`
	ChainRateIdeal4       float32 `json:"chain_rateideal4"`
	ChainRateIdeal9       float32 `json:"chain_rateideal9"`
	ChainRateIdeal10      float32 `json:"chain_rateideal10"`
	ChainRateIdeal11      float32 `json:"chain_rateideal11"`
	ChainRateIdeal12      float32 `json:"chain_rateideal12"`
	ChainRateIdeal13      float32 `json:"chain_rateideal13"`
	ChainRateIdeal14      float32 `json:"chain_rateideal14"`
	DeviceHardwarePercent float64 `json:"Device Hardware%"`
	NotMatchingWork       int     `json:"no_matching_work"`
	ChainAcn2             int     `json:"chain_acn2"`
	ChainAcn3             int     `json:"chain_acn3"`
	ChainAcn4             int     `json:"chain_acn4"`
	ChainAcn9             int     `json:"chain_acn9"`
	ChainAcn10            int     `json:"chain_acn10"`
	ChainAcn11            int     `json:"chain_acn11"`
	ChainAcn12            int     `json:"chain_acn12"`
	ChainAcn13            int     `json:"chain_acn13"`
	ChainAcn14            int     `json:"chain_acn14"`
	ChainAcs2             string  `json:"chain_acs2"`
	ChainAcs3             string  `json:"chain_acs3"`
	ChainAcs4             string  `json:"chain_acs4"`
	ChainAcs9             string  `json:"chain_acs9"`
	ChainAcs10            string  `json:"chain_acs10"`
	ChainAcs11            string  `json:"chain_acs11"`
	ChainAcs12            string  `json:"chain_acs12"`
	ChainAcs13            string  `json:"chain_acs13"`
	ChainAcs14            string  `json:"chain_acs14"`
	ChainHW2              int     `json:"chain_hw2"`
	ChainHW3              int     `json:"chain_hw3"`
	ChainHW4              int     `json:"chain_hw4"`
	ChainHW9              int     `json:"chain_hw9"`
	ChainHW10             int     `json:"chain_hw10"`
	ChainHW11             int     `json:"chain_hw11"`
	ChainHW12             int     `json:"chain_hw12"`
	ChainHW13             int     `json:"chain_hw13"`
	ChainHW14             int     `json:"chain_hw14"`
	ChainRate2            float32 `json:"chain_rate2"`
	ChainRate3            float32 `json:"chain_rate3"`
	ChainRate4            float32 `json:"chain_rate4"`
	ChainRate9            float32 `json:"chain_rate9"`
	ChainRate10           float32 `json:"chain_rate10"`
	ChainRate11           float32 `json:"chain_rate11"`
	ChainRate12           float32 `json:"chain_rate12"`
	ChainRate13           float32 `json:"chain_rate13"`
	ChainRate14           float32 `json:"chain_rate14"`
	ChainXtime2           string  `json:"chain_xtime2"`
	ChainXtime3           string  `json:"chain_xtime3"`
	ChainXtime4           string  `json:"chain_xtime4"`
	ChainXtime9           string  `json:"chain_xtime9"`
	ChainXtime10          string  `json:"chain_xtime10"`
	ChainXtime11          string  `json:"chain_xtime11"`
	ChainXtime12          string  `json:"chain_xtime12"`
	ChainXtime13          string  `json:"chain_xtime13"`
	ChainXtime14          string  `json:"chain_xtime14"`
	ChainOffside2         Number  `json:"chain_offside_2"`
	ChainOffside3         Number  `json:"chain_offside_3"`
	ChainOffside4         Number  `json:"chain_offside_4"`
	ChainOffside9         Number  `json:"chain_offside_9"`
	ChainOffside10        Number  `json:"chain_offside_10"`
	ChainOffside11        Number  `json:"chain_offside_11"`
	ChainOffside12        Number  `json:"chain_offside_12"`
	ChainOffside13        Number  `json:"chain_offside_13"`
	ChainOffside14        Number  `json:"chain_offside_14"`
	ChainOpenCore2        Number  `json:"chain_opencore_2"`
	ChainOpenCore3        Number  `json:"chain_opencore_3"`
	ChainOpenCore4        Number  `json:"chain_opencore_4"`
	ChainOpenCore9        Number  `json:"chain_opencore_9"`
	ChainOpenCore10       Number  `json:"chain_opencore_10"`
	ChainOpenCore11       Number  `json:"chain_opencore_11"`
	ChainOpenCore12       Number  `json:"chain_opencore_12"`
	ChainOpenCore13       Number  `json:"chain_opencore_13"`
	ChainOpenCore14       Number  `json:"chain_opencore_14"`
	MinerID               string  `json:"miner_id"`
	MinerVersion          string  `json:"miner_version"`
}

// StatsD3 -  antminer stats related to D3
type StatsD3 struct {
	Stats                 int    `json:"STATS"`
	CGMiner               string `json:"CGMiner,omitempty"`
	Miner                 string
	CompileTime           string
	Type                  string
	ID                    string
	Elapsed               int64 `json:"Elapsed"`
	Calls                 int
	Wait                  float64
	Min                   float64
	Max                   float64
	Ghs5s                 Number  `json:"GHS 5s"`
	GhsAverage            float64 `json:"GHS av"`
	MinerCount            int16   `json:"miner_count"`
	Frequency             Number  `json:"frequency"`
	FanNum                int16   `json:"fan_num"`
	Fan1                  int16   `json:"fan1"`
	Fan2                  int16   `json:"fan2"`
	DeviceHardwarePercent float64 `json:"Device Hardware%"`
	NotMatchingWork       int     `json:"no_matching_work"`
	TempNum               int16   `json:"temp_num"`
	Temp1                 int16   `json:"temp1"`
	Temp2                 int16   `json:"temp2"`
	Temp3                 int16   `json:"temp3"`
	Temp4                 int16   `json:"temp4"`
	Temp2_1               int16   `json:"temp2_1"`
	Temp2_2               int16   `json:"temp2_2"`
	Temp2_3               int16   `json:"temp2_3"`
	TempMax               int16   `json:"temp_max"`
	ChainAcn1             int     `json:"chain_acn1"`
	ChainAcn2             int     `json:"chain_acn2"`
	ChainAcn3             int     `json:"chain_acn3"`
	ChainAcs1             string  `json:"chain_acs1"`
	ChainAcs2             string  `json:"chain_acs2"`
	ChainAcs3             string  `json:"chain_acs3"`
	ChainHW1              int     `json:"chain_hw1"`
	ChainHW2              int     `json:"chain_hw2"`
	ChainHW3              int     `json:"chain_hw3"`
	ChainRate1            float32 `json:"chain_rate1"`
	ChainRate2            float32 `json:"chain_rate2"`
	ChainRate3            float32 `json:"chain_rate3"`
}

// StatsL3 - antminer stats related to L3+
type StatsL3 struct {
	CGMiner               string
	Miner                 string
	CompileTime           string
	Type                  string
	MinerCount            int16 `json:"miner_count"`
	Elapsed               int64 `json:"Elapsed"`
	Wait                  float64
	DeviceHardwarePercent float64 `json:"Device Hardware%"`
	NotMatchingWork       int     `json:"no_matching_work"`
	Stats                 int     `json:"STATS"`
	Min                   float64
	Max                   float64
	ID                    string
	Calls                 int
	Frequency             Number  `json:"frequency"`
	FanNum                int16   `json:"fan_num"`
	Fan1                  int16   `json:"fan1"`
	Fan2                  int16   `json:"fan2"`
	TempMax               int16   `json:"temp_max"`
	TempNum               int16   `json:"temp_num"`
	Temp1                 int16   `json:"temp1"`
	Temp2                 int16   `json:"temp2"`
	Temp3                 int16   `json:"temp3"`
	Temp4                 int16   `json:"temp4"`
	Temp2_1               int16   `json:"temp2_1"`
	Temp2_2               int16   `json:"temp2_2"`
	Temp2_3               int16   `json:"temp2_3"`
	Temp2_4               int16   `json:"temp2_4"`
	Temp3_1               int16   `json:"temp31"`
	Temp3_2               int16   `json:"temp32"`
	Temp3_3               int16   `json:"temp33"`
	Temp3_4               int16   `json:"temp34"`
	Temp4_1               int16   `json:"temp4_1"`
	Temp4_2               int16   `json:"temp4_2"`
	Temp4_3               int16   `json:"temp4_3"`
	Temp4_4               int16   `json:"temp4_4"`
	Ghs5s                 Number  `json:"GHS 5s"`
	GhsAverage            float64 `json:"GHS av"`
	ChainHW1              int     `json:"chain_hw1"`
	ChainHW2              int     `json:"chain_hw2"`
	ChainHW3              int     `json:"chain_hw3"`
	ChainHW4              int     `json:"chain_hw4"`
	ChainAcs1             string  `json:"chain_acs1"`
	ChainAcs2             string  `json:"chain_acs2"`
	ChainAcs3             string  `json:"chain_acs3"`
	ChainAcs4             string  `json:"chain_acs4"`
	ChainAcn1             int     `json:"chain_acn1"`
	ChainAcn2             int     `json:"chain_acn2"`
	ChainAcn3             int     `json:"chain_acn3"`
	ChainAcn4             int     `json:"chain_acn4"`
	ChainRate1            float32 `json:"chain_rate1"`
	ChainRate2            float32 `json:"chain_rate2"`
	ChainRate3            float32 `json:"chain_rate3"`
	ChainRate4            float32 `json:"chain_rate4"`
}

// StatsS9 - get stats from antminer S9
type StatsS9 struct {
	BMMiner               string
	API                   string
	Miner                 string
	CompileTime           string
	Type                  string
	MinerID               string `json:"miner_id"`
	MinerVersion          string `json:"miner_version"`
	MinerCount            int16  `json:"miner_count"`
	Elapsed               int64  `json:"Elapsed"`
	Wait                  float64
	DeviceHardwarePercent float64 `json:"Device Hardware%"`
	Stats                 int     `json:"STATS"`
	Max                   float64
	NotMatchingWork       int `json:"no_matching_work"`
	ID                    string
	Calls                 int
	Min                   float64
	TotalAcn              int16   `json:"total_acn"`
	TotalRate             float32 `json:"total_rate"`
	TotalRateIdeal        float32 `json:"total_rateideal"`
	TotalFrequencyAvg     float32 `json:"total_freqavg"`
	Frequency             Number  `json:"frequency"`
	FrequencyAvg6         float32 `json:"freq_avg6"`
	FrequencyAvg7         float32 `json:"freq_avg7"`
	FrequencyAvg8         float32 `json:"freq_avg8"`
	FanNum                int16   `json:"fan_num"`
	Fan3                  int16   `json:"fan3"`
	Fan6                  int16   `json:"fan6"`
	TempMax               int16   `json:"temp_max"`
	TempNum               int16   `json:"temp_num"`
	Temp6                 int16   `json:"temp6"`
	Temp7                 int16   `json:"temp7"`
	Temp8                 int16   `json:"temp8"`
	Temp2_6               int16   `json:"temp2_6"`
	Temp2_7               int16   `json:"temp2_7"`
	Temp2_8               int16   `json:"temp2_8"`
	Ghs5s                 Number  `json:"GHS 5s"`
	GhsAverage            float64 `json:"GHS av"`
	ChainHW6              int     `json:"chain_hw6"`
	ChainHW7              int     `json:"chain_hw7"`
	ChainHW8              int     `json:"chain_hw8"`
	ChainAcs6             string  `json:"chain_acs6"`
	ChainAcs7             string  `json:"chain_acs7"`
	ChainAcs8             string  `json:"chain_acs8"`
	ChainAcn6             int     `json:"chain_acn6"`
	ChainAcn7             int     `json:"chain_acn7"`
	ChainAcn8             int     `json:"chain_acn8"`
	ChainRate6            float32 `json:"chain_rate6"`
	ChainRate7            float32 `json:"chain_rate7"`
	ChainRate8            float32 `json:"chain_rate8"`
	ChainRateIdeal6       float32 `json:"chain_rateideal6"`
	ChainRateIdeal7       float32 `json:"chain_rateideal7"`
	ChainRateIdeal8       float32 `json:"chain_rateideal8"`
	ChainOpenCore6        Number  `json:"chain_opencore_6"`
	ChainOpenCore7        Number  `json:"chain_opencore_7"`
	ChainOpenCore8        Number  `json:"chain_opencore_8"`
	ChainOffside6         Number  `json:"chain_offside_6"`
	ChainOffside7         Number  `json:"chain_offside_7"`
	ChainOffside8         Number  `json:"chain_offside_8"`
	// ChainXtime6 need to be parsed?
	ChainXtime6 string `json:"chain_xtime6"`
	ChainXtime7 string `json:"chain_xtime7"`
	ChainXtime8 string `json:"chain_xtime8"`
}

// Summary - miner summary
type Summary struct {
	Accepted              int64
	BestShare             int64   `json:"Best Share"`
	DeviceHardwarePercent float64 `json:"Device Hardware%"`
	DeviceRejectedPercent float64 `json:"Device Rejected%"`
	DifficultyAccepted    float64 `json:"Difficulty Accepted"`
	DifficultyRejected    float64 `json:"Difficulty Rejected"`
	DifficultyStale       float64 `json:"Difficulty Stale"`
	Discarded             int64
	Elapsed               int64
	FoundBlocks           int64 `json:"Found Blocks"`
	GetFailures           int64 `json:"Get Failures"`
	Getworks              int64
	HardwareErrors        int64 `json:"Hardware Errors"`
	LocalWork             int64 `json:"Local Work"`
	// non s7/s9/d3 etc.
	MHS5s float64 `json:"MHS 5s"`
	MHSav float64 `json:"MHS av"`
	// s7 and later
	GHS5s               Number  `json:"GHS 5s"`
	GHSav               float64 `json:"GHS av"`
	NetworkBlocks       int64   `json:"Network Blocks"`
	PoolRejectedPercent float64 `json:"Pool Rejected%"`
	PoolStalePercent    float64 `json:"Pool Stale%"`
	Rejected            int64
	RemoteFailures      int64 `json:"Remote Failures"`
	Stale               int64
	TotalMH             float64 `json:"Total MH"`
	Utility             float64
	WorkUtility         float64 `json:"Work Utility"`
	LastGetWork         int     `json:"Last getwork"`
}

// Devs - device info
type Devs struct {
	GPU                 int64
	Enabled             string
	Status              string
	Temperature         float64
	FanSpeed            int     `json:"Fan Speed"`
	FanPercent          int64   `json:"Fan Percent"`
	GPUClock            int64   `json:"GPU Clock"`
	MemoryClock         int64   `json:"Memory Clock"`
	GPUVoltage          float64 `json:"GPU Voltage"`
	Powertune           int64
	MHSav               float64 `json:"MHS av"`
	MHS5s               float64 `json:"MHS 5s"`
	Accepted            int64
	Rejected            int64
	HardwareErrors      int64 `json:"Hardware Errors"`
	Utility             float64
	Intensity           string
	LastSharePool       int64   `json:"Last Share Pool"`
	LashShareTime       int64   `json:"Lash Share Time"`
	TotalMH             float64 `json:"TotalMH"`
	Diff1Work           int64   `json:"Diff1 Work"`
	DifficultyAccepted  float64 `json:"Difficulty Accepted"`
	DifficultyRejected  float64 `json:"Difficulty Rejected"`
	LastShareDifficulty float64 `json:"Last Share Difficulty"`
	LastValidWork       int64   `json:"Last Valid Work"`
	DeviceHardware      float64 `json:"Device Hardware%"`
	DeviceRejected      float64 `json:"Device Rejected%"`
	DeviceElapsed       int64   `json:"Device Elapsed"`
}

// Pool - get working pools
type Pool struct {
	Accepted            int64
	BestShare           Number  `json:"Best Share"`
	Diff1Shares         int64   `json:"Diff1 Shares"`
	DifficultyAccepted  float64 `json:"Difficulty Accepted"`
	DifficultyRejected  float64 `json:"Difficulty Rejected"`
	DifficultyStale     float64 `json:"Difficulty Stale"`
	Discarded           int64
	GetFailures         int64 `json:"Get Failures"`
	Getworks            int64
	HasGBT              bool    `json:"Has GBT"`
	HasStratum          bool    `json:"Has Stratum"`
	LastShareDifficulty float64 `json:"Last Share Difficulty"`
	LastShareTime       string  `json:"Last Share Time"`
	LongPoll            string  `json:"Long Poll"`
	Pool                int64   `json:"POOL"`
	PoolRejectedPercent float64 `json:"Pool Rejected%"`
	PoolStalePercent    float64 `json:"Pool Stale%"`
	Priority            int64
	ProxyType           string `json:"Proxy Type"`
	Proxy               string
	Quota               int64
	Rejected            int64
	RemoteFailures      int64 `json:"Remote Failures"`
	Stale               int64
	Status              string
	StratumActive       bool   `json:"Stratum Active"`
	StratumURL          string `json:"Stratum URL"`
	URL                 string `json:"URL"`
	User                string `json:"User"`
	Works               int64
}

type statsResponse struct {
	GenericResponse
	Stats []GenericStats `json:"STATS"`
}

type summaryResponse struct {
	GenericResponse
	Summary []Summary `json:"SUMMARY"`
}

type devsResponse struct {
	GenericResponse
	Devs []Devs `json:"DEVS"`
}

type poolsResponse struct {
	GenericResponse
	Pools []Pool `json:"POOLS"`
}
