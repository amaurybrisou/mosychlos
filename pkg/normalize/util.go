package normalize

import (
	"sort"
	"time"
)

// toUTC converts any time to UTC (no-op if already UTC).
func toUTC(t time.Time) time.Time {
	return t.UTC()
}

// epochToUTC converts Unix epoch seconds to UTC time.
func epochToUTC(sec int64) time.Time {
	return time.Unix(sec, 0).UTC()
}

// sortPointsAsc sorts timeseries points by time ascending.
func sortPointsAsc(points []TimeseriesPoint) {
	sort.Slice(points, func(i, j int) bool { return points[i].T.Before(points[j].T) })
}
