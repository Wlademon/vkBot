package time

import "time"

var Location *time.Location

func InitTime(locationStr string) *time.Location {
	location, err := time.LoadLocation(locationStr)
	if err != nil {
		panic("Time not set.")
	}
	Location = location

	return Location
}

func Parse(format string, value string) (time.Time, error) {
	parse, err := time.Parse(format, value)
	if err != nil {
		return parse, err
	}

	return parse.In(Location), nil
}

func Now() time.Time {
	return time.Now().In(Location)
}
