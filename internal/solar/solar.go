package solar

import (
	"errors"
	"math"
	"time"
)

var ErrNoEvent = errors.New("solar event does not occur on this date")

const (
	officialZenith     = 90.833
	civilZenith        = 96.0
	nauticalZenith     = 102.0
	astronomicalZenith = 108.0
)

func Sunrise(date time.Time, lat, lon float64, loc *time.Location) (time.Time, error) {
	return event(date, lat, lon, loc, officialZenith, true)
}

func Sunset(date time.Time, lat, lon float64, loc *time.Location) (time.Time, error) {
	return event(date, lat, lon, loc, officialZenith, false)
}

func CivilDawn(date time.Time, lat, lon float64, loc *time.Location) (time.Time, error) {
	return event(date, lat, lon, loc, civilZenith, true)
}

func CivilDusk(date time.Time, lat, lon float64, loc *time.Location) (time.Time, error) {
	return event(date, lat, lon, loc, civilZenith, false)
}

func NauticalDawn(date time.Time, lat, lon float64, loc *time.Location) (time.Time, error) {
	return event(date, lat, lon, loc, nauticalZenith, true)
}

func NauticalDusk(date time.Time, lat, lon float64, loc *time.Location) (time.Time, error) {
	return event(date, lat, lon, loc, nauticalZenith, false)
}

func AstronomicalDawn(date time.Time, lat, lon float64, loc *time.Location) (time.Time, error) {
	return event(date, lat, lon, loc, astronomicalZenith, true)
}

func AstronomicalDusk(date time.Time, lat, lon float64, loc *time.Location) (time.Time, error) {
	return event(date, lat, lon, loc, astronomicalZenith, false)
}

func SolarNoon(date time.Time, lat, lon float64, loc *time.Location) (time.Time, error) {
	rise, riseErr := Sunrise(date, lat, lon, loc)
	set, setErr := Sunset(date, lat, lon, loc)
	if riseErr == nil && setErr == nil {
		return rise.Add(set.Sub(rise) / 2), nil
	}
	if loc == nil {
		loc = time.Local
	}
	localDate := date.In(loc)
	year, month, day := localDate.Date()
	return time.Date(year, month, day, 12, 0, 0, 0, loc), nil
}

func event(date time.Time, lat, lon float64, loc *time.Location, zenith float64, rise bool) (time.Time, error) {
	if loc == nil {
		loc = time.Local
	}
	localDate := date.In(loc)
	year, month, day := localDate.Date()
	n := float64(localDate.YearDay())
	lngHour := lon / 15

	var t float64
	if rise {
		t = n + ((6 - lngHour) / 24)
	} else {
		t = n + ((18 - lngHour) / 24)
	}

	meanAnomaly := normalizeDegrees((0.9856 * t) - 3.289)
	trueLong := normalizeDegrees(meanAnomaly + (1.916 * sinDeg(meanAnomaly)) + (0.020 * sinDeg(2*meanAnomaly)) + 282.634)
	rightAscension := normalizeDegrees(atanDeg(0.91764 * tanDeg(trueLong)))

	lQuadrant := math.Floor(trueLong/90) * 90
	raQuadrant := math.Floor(rightAscension/90) * 90
	rightAscension = (rightAscension + lQuadrant - raQuadrant) / 15

	sinDec := 0.39782 * sinDeg(trueLong)
	cosDec := math.Cos(math.Asin(sinDec))
	cosH := (cosDeg(zenith) - (sinDec * sinDeg(lat))) / (cosDec * cosDeg(lat))
	if cosH > 1 || cosH < -1 {
		return time.Time{}, ErrNoEvent
	}

	var hourAngle float64
	if rise {
		hourAngle = 360 - acosDeg(cosH)
	} else {
		hourAngle = acosDeg(cosH)
	}
	hourAngle /= 15

	localMean := hourAngle + rightAscension - (0.06571 * t) - 6.622
	utcHour := normalizeHours(localMean - lngHour)
	hour := int(math.Floor(utcHour))
	minuteFloat := (utcHour - float64(hour)) * 60
	minute := int(math.Floor(minuteFloat))
	second := int(math.Round((minuteFloat - float64(minute)) * 60))
	if second == 60 {
		second = 0
		minute++
	}
	if minute == 60 {
		minute = 0
		hour++
	}
	if hour == 24 {
		hour = 0
	}

	utcClock := time.Date(year, month, day, hour, minute, second, 0, time.UTC)
	for _, dayOffset := range []int{0, -1, 1} {
		candidate := utcClock.AddDate(0, 0, dayOffset).In(loc)
		candidateYear, candidateMonth, candidateDay := candidate.Date()
		if candidateYear == year && candidateMonth == month && candidateDay == day {
			return candidate, nil
		}
	}
	return utcClock.In(loc), nil
}

func normalizeDegrees(v float64) float64 {
	for v < 0 {
		v += 360
	}
	for v >= 360 {
		v -= 360
	}
	return v
}

func normalizeHours(v float64) float64 {
	for v < 0 {
		v += 24
	}
	for v >= 24 {
		v -= 24
	}
	return v
}

func sinDeg(v float64) float64 {
	return math.Sin(v * math.Pi / 180)
}

func cosDeg(v float64) float64 {
	return math.Cos(v * math.Pi / 180)
}

func tanDeg(v float64) float64 {
	return math.Tan(v * math.Pi / 180)
}

func atanDeg(v float64) float64 {
	return math.Atan(v) * 180 / math.Pi
}

func acosDeg(v float64) float64 {
	return math.Acos(v) * 180 / math.Pi
}
