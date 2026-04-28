# Weather Widget Server - Roadmap

## Current Status

The server infrastructure is in place with basic routing and external API integrations:

- **Geocoding**: OpenWeatherMap Geocoding API for location lookup (city/state → lat/lon)
- **Weather Stations**: NOAA CDO API for retrieving nearby weather stations
- **CLI**: Configurable host, port, and API keys via flags/environment variables
- **Testing & Build**: Makefile with test coverage requirements (50% minimum), Docker build support

## Completed

- [x] Basic HTTP server with routes
- [x] Location geocoding integration
- [x] Weather station lookup
- [x] Command-line argument parsing
- [x] API key configuration

## Outstanding TODOs

### High Priority

1. **Complete weather data retrieval flow**
   - After finding a station, fetch actual weather/forecast data from NOAA
   - Decide on response schema for weather data

2. **Historical average temperatures**
   - Research and integrate a data source for historical averages by location
   - Consider NOAA's available datasets or alternative sources

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

8. **Request validation & sanitization** (currently marked as TODO in code)
9. **Response caching** (for frequently requested locations)
10. **Rate limiting** (for API protection)

## Outstanding Design Decisions

- **Response Schema**: What fields should the `/weather` endpoint return? (current forecast, historical avg, metadata?)
- **Location Disambiguation**: When geocoding returns multiple results, should server return options or auto-select?
- **Weather Data Source**: Continue with NOAA CDO, or explore alternatives?
- **Historical Data Source**: Which service for historical temperature averages?
- **Deployment**: Container-first (Dockerfile exists), serverless, traditional VPS?

## Notes

- Project uses Go 1.26.2
- Minimum test coverage enforced at 50%
- Consider adding staticcheck linting to CI/CD
