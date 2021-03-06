package time

import "time"

var Location *time.Location

func InitTime(offset time.Duration) *time.Location {
	Location = time.FixedZone("Current", int(offset))

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
