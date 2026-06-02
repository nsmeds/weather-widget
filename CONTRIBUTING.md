# Contributing

## Prerequisites

- [Go](https://go.dev/doc/install) 1.22+
- [Docker](https://docs.docker.com/get-docker/) (for container builds)
- [Make](https://www.gnu.org/software/make/) (typically pre-installed on macOS/Linux)
- [staticcheck](https://staticcheck.dev) (`go install honnef.co/go/tools/cmd/staticcheck@latest`) for linting

## Common commands

```bash
make test    # run tests with coverage (minimum 65% enforced)
make lint    # run staticcheck
make build   # build Docker image (linux/amd64)
```

## Running locally

The server requires two API keys, available as flags or environment variables:

| Flag | Env var | Source |
|---|---|---|
| `-geocode-api-key` | `OPEN_WEATHER_API_KEY` | [OpenWeatherMap](https://openweathermap.org/api/geocoding-api) |
| `-weather-api-key` | `NCDC_API_KEY` | [NOAA CDO](https://www.ncdc.noaa.gov/cdo-web/token) |

```bash
export OPEN_WEATHER_API_KEY=your_key
export NCDC_API_KEY=your_key

go run ./cmd
```

The server starts on `localhost:8080` by default. Use `-host` and `-port` to override.

## Querying the weather endpoint

Send a `POST` to `/weather` with a US city and state as the request body:

```bash
curl -s -X POST http://localhost:8080/weather -d "Boston, MA" | jq
```

If geocoding returns multiple matches (e.g. a common city name), the server responds with the list of candidates so the caller can disambiguate. Otherwise it returns the full weather response:

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
