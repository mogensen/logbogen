package services

import (
	"log/slog"
	"time"

	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/types"
)

const dayLayout = "2006-01-02"

// WeatherViewModel is the weather as the activity page shows it
type WeatherViewModel struct {
	Temperature float64
	Icon        string
}

func isToday(date time.Time) bool {
	return date.Format(dayLayout) == time.Now().Format(dayLayout)
}

// isPastDay reports whether the day is over, which is when its weather no longer changes
func isPastDay(date time.Time) bool {
	return date.Format(dayLayout) < time.Now().Format(dayLayout)
}

// lookupWeather asks the weather provider, a failed lookup is logged and reported as missing
func (s *ActivityService) lookupWeather(lat, lng float64, date time.Time) (*types.Weather, bool) {
	if s.weatherSvc == nil {
		return nil, false
	}

	weather, err := s.weatherSvc.GetWeather(lat, lng, date)
	if err != nil {
		slog.Error("Error getting weather data", "error", err)
		return nil, false
	}

	temperature, code, ok := weather.Day()
	if !ok {
		slog.Warn("Weather data is not available yet", "date", date.Format(dayLayout))
		return nil, false
	}
	return &types.Weather{Temperature: temperature, Code: code}, true
}

// pastWeather returns the weather to store for a day that is over, and nothing for today or the future
func (s *ActivityService) pastWeather(lat, lng float64, date time.Time) *types.Weather {
	if !isPastDay(date) {
		return nil
	}
	weather, _ := s.lookupWeather(lat, lng, date)
	return weather
}

// setWeather puts the weather on an activity that is about to be saved
func setWeather(activity *dal.Activity, weather *types.Weather) {
	if weather == nil {
		return
	}
	activity.WeatherTemperature = &weather.Temperature
	activity.WeatherCode = &weather.Code
}

// weatherForView returns the weather to show: stored weather as is, today's looked up live,
// and a past day without stored weather looked up once and stored.
func (s *ActivityService) weatherForView(activity *dal.Activity, userID uint64) *types.Weather {
	if activity.WeatherTemperature != nil && activity.WeatherCode != nil {
		return &types.Weather{Temperature: *activity.WeatherTemperature, Code: *activity.WeatherCode}
	}

	switch {
	case isToday(activity.Date):
		weather, _ := s.lookupWeather(activity.Lat, activity.Lng, activity.Date)
		return weather
	case isPastDay(activity.Date):
		weather, ok := s.lookupWeather(activity.Lat, activity.Lng, activity.Date)
		if !ok {
			return nil
		}
		activityID := activity.ID.String()
		if err := s.activityDal.UpdateWeather(activityID, userID, &weather.Temperature, &weather.Code); err != nil {
			slog.Error("Error storing weather data", "error", err)
		}
		return weather
	}
	return nil
}

func weatherViewModel(weather *types.Weather) *WeatherViewModel {
	if weather == nil {
		return nil
	}
	return &WeatherViewModel{Temperature: weather.Temperature, Icon: GetWeatherIcon(weather.Code)}
}

// placeOrDayChanged reports whether the new date or coordinates make the stored weather wrong
func placeOrDayChanged(existing *dal.Activity, date time.Time, lat, lng float64) bool {
	return existing.Date.Format(dayLayout) != date.Format(dayLayout) || existing.Lat != lat || existing.Lng != lng
}
