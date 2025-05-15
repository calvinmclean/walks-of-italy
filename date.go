package tours

import (
	"fmt"
	"time"
)

type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func NewDate(year int, month time.Month, day int) Date {
	return Date{year, month, day}
}

func (d Date) Add(years, months, days int) Date {
	return DateFromTime(d.ToTime().AddDate(years, months, days))
}

func (d Date) ToTime() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}

func DateFromTime(t time.Time) Date {
	return Date{
		Year:  t.Year(),
		Month: t.Month(),
		Day:   t.Day(),
	}
}

func (d Date) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`"%04d-%02d-%02d"`, d.Year, d.Month, d.Day)
	return []byte(s), nil
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}
