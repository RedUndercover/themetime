package solar

import (
	"testing"
	"time"
)

func TestEventsStayOnRequestedLocalDateAcrossUTCBoundaries(t *testing.T) {
	tests := []struct {
		name string
		zone string
		lat  float64
		lon  float64
	}{
		{name: "New York sunset crosses into next UTC date", zone: "America/New_York", lat: 40.7128, lon: -74.0060},
		{name: "Tokyo sunrise comes from previous UTC date", zone: "Asia/Tokyo", lat: 35.6762, lon: 139.6503},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loc, err := time.LoadLocation(test.zone)
			if err != nil {
				t.Fatal(err)
			}
			date := time.Date(2026, time.July, 20, 12, 0, 0, 0, loc)
			for name, eventFn := range map[string]func(time.Time, float64, float64, *time.Location) (time.Time, error){
				"sunrise": Sunrise,
				"sunset":  Sunset,
			} {
				at, err := eventFn(date, test.lat, test.lon, loc)
				if err != nil {
					t.Fatalf("%s: %v", name, err)
				}
				year, month, day := at.In(loc).Date()
				if year != 2026 || month != time.July || day != 20 {
					t.Fatalf("%s resolved to %s, want local date 2026-07-20", name, at.In(loc).Format(time.RFC3339))
				}
			}
		})
	}
}

func TestSolarNoonFallsBetweenSameDaySunriseAndSunset(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	date := time.Date(2026, time.July, 20, 12, 0, 0, 0, loc)
	rise, err := Sunrise(date, 40.7128, -74.0060, loc)
	if err != nil {
		t.Fatal(err)
	}
	set, err := Sunset(date, 40.7128, -74.0060, loc)
	if err != nil {
		t.Fatal(err)
	}
	noon, err := SolarNoon(date, 40.7128, -74.0060, loc)
	if err != nil {
		t.Fatal(err)
	}

	if !rise.Before(noon) || !noon.Before(set) {
		t.Fatalf("events are out of order: sunrise=%s noon=%s sunset=%s", rise, noon, set)
	}
	want := rise.Add(set.Sub(rise) / 2)
	if !noon.Equal(want) {
		t.Fatalf("solar noon=%s, want midpoint %s", noon, want)
	}
	if noon.Hour() < 11 || noon.Hour() > 14 {
		t.Fatalf("solar noon clock=%s, want a daytime hour", noon.Format("15:04:05"))
	}
}
