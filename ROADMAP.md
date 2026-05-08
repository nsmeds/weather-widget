# Weather Widget Server - Roadmap

## Primary Use Case

For a given US location (city/state), return:
- **Current conditions**: temperature, weather description, humidity, wind speed
- **Today's forecast**: high/low temperatures and short forecast text
- **Historical averages**: 30-year climate normals for today's date (high/low temperatures)

Geographic scope: US-only for now. International expansion can be revisited if the project grows.

## Current Status

The server infrastructure is in place with basic routing and external API integrations:

- **Geocoding**: OpenWeatherMap Geocoding API for location lookup (city/state → lat/lon)
- **Weather Stations**: NOAA CDO API for retrieving nearby GHCND weather stations
- **CLI**: Configurable host, port, and API keys via flags/environment variables
- **Testing & Build**: Makefile with test coverage requirements (65% minimum), Docker build support

## Completed

- [x] Basic HTTP server with routes
- [x] Location geocoding integration
- [x] Weather station lookup (NOAA CDO/GHCND)
- [x] Command-line argument parsing
- [x] API key configuration

## Outstanding TODOs

### High Priority

1. **Complete weather data retrieval flow**
   - Integrate NWS API (api.weather.gov) for current conditions and today's forecast
   - Wire up full `/weather` endpoint to return current + forecast + averages

2. **Historical average temperatures**
   - Use NOAA CDO NORMAL_DLY dataset for 30-year climate normals (1981-2010)

3. **Input validation & sanitization**
   - Sanitize geocoding query input
   - Validate API responses and error cases

### Medium Priority

4. **State code handling**
   - Convert 2-letter state abbreviations to 3-letter codes where needed for APIs

5. **Error handling**
   - Improve HTTP error responses with consistent JSON format
   - Add error logging throughout request pipeline

6. **Logging & Observability**
   - Add structured logging (logger TODO in server.go)
   - Add metrics collection (metrics TODO in server.go)

7. **Configuration management**
   - Consider config file support alongside CLI flags

### Nice to Have

8. **Response caching** (for frequently requested locations)
9. **Rate limiting** (for API protection)

## Response Schema

```json
{
  "location": { "name": "Boston, MA", "lat": 42.36, "lon": -71.06 },
  "current": {
    "temperature_f": 65.0,
    "conditions": "Partly Cloudy",
    "humidity": 72.5,
    "wind_speed_mph": 9.2,
    "timestamp": "2026-05-07T14:00:00Z"
  },
  "forecast": {
    "name": "Today",
    "short_forecast": "Mostly Sunny",
    "temperature_high_f": 68.0,
    "temperature_low_f": 45.0
  },
  "averages": {
    "temperature_high_f": 63.0,
    "temperature_low_f": 46.0,
    "period": "1981-2010 climate normals"
  }
}
```

## Design Decisions

- **Weather Data Source**: NWS API (api.weather.gov) — free, no key required, US-only
- **Historical Data Source**: NOAA CDO NORMAL_DLY dataset — 30-year (1981-2010) climate normals via existing NCDC key
- **Location Disambiguation**: When geocoding returns multiple results, return all options for the caller to select
- **Geographic Scope**: US-only for now; may expand to international sources in the future
- **Deployment**: Container-first (Dockerfile exists), serverless, traditional VPS?

## Notes

- Project uses Go 1.26.2
- Minimum test coverage enforced at 65%