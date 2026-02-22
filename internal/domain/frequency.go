package domain

import "fmt"

type Frequency string

const (
	FrequencyDaily     Frequency = "daily"
	FrequencyWeekday   Frequency = "weekday"
	FrequencyWeekly    Frequency = "weekly"
	FrequencyBiweekly  Frequency = "biweekly"
	FrequencyMonthly   Frequency = "monthly"
	FrequencyQuarterly Frequency = "quarterly"
	FrequencyYearly    Frequency = "yearly"
)

var validFrequencies = map[Frequency]bool{
	FrequencyDaily:     true,
	FrequencyWeekday:   true,
	FrequencyWeekly:    true,
	FrequencyBiweekly:  true,
	FrequencyMonthly:   true,
	FrequencyQuarterly: true,
	FrequencyYearly:    true,
}

func (f Frequency) Valid() bool {
	return validFrequencies[f]
}

func (f Frequency) Validate() error {
	if !f.Valid() {
		return fmt.Errorf("%w: invalid frequency %q", ErrValidation, f)
	}
	return nil
}

func AllFrequencies() []Frequency {
	return []Frequency{
		FrequencyDaily, FrequencyWeekday, FrequencyWeekly, FrequencyBiweekly,
		FrequencyMonthly, FrequencyQuarterly, FrequencyYearly,
	}
}
