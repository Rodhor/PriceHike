package domain

type Settings struct {
	CSVDelimiter         CSVDelimiter `json:"csvDelimiter"`
	DefaultDecimalPlaces int          `json:"defaultDecimalPlaces"`
}

type CSVDelimiter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func DefaultSettings() Settings {
	return Settings{
		CSVDelimiter:         CSVDelimiter{Name: "Comma", Value: ","},
		DefaultDecimalPlaces: 2,
	}
}

func (s Settings) Validate() error {
	if s.DefaultDecimalPlaces < 0 || s.DefaultDecimalPlaces > 10 {
		return ErrInvalidSettings
	}

	if s.CSVDelimiter.Value == "" {
		return ErrInvalidSettings
	}
	runes := []rune(s.CSVDelimiter.Value)
	if len(runes) != 1 {
		return ErrInvalidSettings
	}
	if runes[0] == '\n' || runes[0] == '\r' {
		return ErrInvalidSettings
	}
	return nil
}

func (s Settings) Delimiter() rune {
	runes := []rune(s.CSVDelimiter.Value)
	if len(runes) == 0 {
		return ','
	}
	return runes[0]
}
