package i18n

var frequencyNames = map[Locale]map[string]string{
	DE: {
		"daily":     "Täglich",
		"weekday":   "Werktäglich",
		"weekly":    "Wöchentlich",
		"biweekly":  "Zweiwöchentlich",
		"monthly":   "Monatlich",
		"quarterly": "Vierteljährlich",
		"yearly":    "Jährlich",
	},
	EN: {
		"daily":     "Daily",
		"weekday":   "Weekday",
		"weekly":    "Weekly",
		"biweekly":  "Biweekly",
		"monthly":   "Monthly",
		"quarterly": "Quarterly",
		"yearly":    "Yearly",
	},
}

// FrequencyName returns the localized display name for a frequency.
func (b *Bundle) FrequencyName(locale Locale, freq string) string {
	if names, ok := frequencyNames[locale]; ok {
		if name, ok := names[freq]; ok {
			return name
		}
	}
	// Fallback to default locale
	if names, ok := frequencyNames[b.defaultLocale]; ok {
		if name, ok := names[freq]; ok {
			return name
		}
	}
	return freq
}
