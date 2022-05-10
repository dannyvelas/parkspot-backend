package config

type Constants struct {
	dateFormat string
}

const dateFormat = "2006-01-02"

func newConstants() Constants {
	return Constants{
		dateFormat: dateFormat,
	}
}

func (constants Constants) DateFormat() string {
	return constants.dateFormat
}
