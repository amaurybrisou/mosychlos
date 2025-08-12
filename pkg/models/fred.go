package models

// FRED (Federal Reserve Economic Data) API Response Types

// FREDSeries represents a FRED economic data series
type FREDSeries struct {
	ID                      string `json:"id"`
	RealtimeStart           string `json:"realtime_start"`
	RealtimeEnd             string `json:"realtime_end"`
	Title                   string `json:"title"`
	ObservationStart        string `json:"observation_start"`
	ObservationEnd          string `json:"observation_end"`
	Frequency               string `json:"frequency"`
	FrequencyShort          string `json:"frequency_short"`
	Units                   string `json:"units"`
	UnitsShort              string `json:"units_short"`
	SeasonalAdjustment      string `json:"seasonal_adjustment"`
	SeasonalAdjustmentShort string `json:"seasonal_adjustment_short"`
	LastUpdated             string `json:"last_updated"`
	Popularity              int    `json:"popularity"`
	GroupPopularity         int    `json:"group_popularity"`
	Notes                   string `json:"notes"`
}

// FREDSeriesResponse represents FRED series API response
type FREDSeriesResponse struct {
	RealtimeStart string       `json:"realtime_start"`
	RealtimeEnd   string       `json:"realtime_end"`
	OrderBy       string       `json:"order_by"`
	SortOrder     string       `json:"sort_order"`
	Count         int          `json:"count"`
	Offset        int          `json:"offset"`
	Limit         int          `json:"limit"`
	Seriess       []FREDSeries `json:"seriess"`
}

// FREDObservation represents a data observation
type FREDObservation struct {
	RealtimeStart string `json:"realtime_start"`
	RealtimeEnd   string `json:"realtime_end"`
	Date          string `json:"date"`
	Value         string `json:"value"`
}

// FREDObservationsResponse represents FRED observations API response
type FREDObservationsResponse struct {
	RealtimeStart    string            `json:"realtime_start"`
	RealtimeEnd      string            `json:"realtime_end"`
	ObservationStart string            `json:"observation_start"`
	ObservationEnd   string            `json:"observation_end"`
	Units            string            `json:"units"`
	OutputType       int               `json:"output_type"`
	FileType         string            `json:"file_type"`
	OrderBy          string            `json:"order_by"`
	SortOrder        string            `json:"sort_order"`
	Count            int               `json:"count"`
	Offset           int               `json:"offset"`
	Limit            int               `json:"limit"`
	Observations     []FREDObservation `json:"observations"`
}

// FREDSeriesObservations represents series with observations
type FREDSeriesObservations struct {
	SeriesID     string            `json:"series_id"`
	Observations []FREDObservation `json:"observations"`
	Count        int               `json:"count"`
	Offset       int               `json:"offset"`
	Limit        int               `json:"limit"`
	OrderBy      string            `json:"order_by"`
	SortOrder    string            `json:"sort_order"`
}

// FREDCategory represents a FRED category
type FREDCategory struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID int    `json:"parent_id"`
}

// FREDCategoriesResponse represents FRED categories API response
type FREDCategoriesResponse struct {
	Categories []FREDCategory `json:"categories"`
}

// FREDCategories represents categories collection
type FREDCategories struct {
	Categories []FREDCategory `json:"categories"`
}

// FREDCategorySeries represents series within a category
type FREDCategorySeries struct {
	CategoryID string       `json:"category_id"`
	Series     []FREDSeries `json:"series"`
	Count      int          `json:"count"`
	Offset     int          `json:"offset"`
	Limit      int          `json:"limit"`
}

// FREDSearchResults represents search results
type FREDSearchResults struct {
	SearchText string       `json:"search_text"`
	Series     []FREDSeries `json:"series"`
	Count      int          `json:"count"`
	Offset     int          `json:"offset"`
	Limit      int          `json:"limit"`
}

// FRED Error response
type FREDError struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

// FREDGeoRegionalData represents FRED GeoFRED regional data API response
type FREDGeoRegionalData struct {
	Meta FREDGeoMeta `json:"meta"`
}

// FREDGeoMeta represents the meta section of GeoFRED response
type FREDGeoMeta struct {
	Data map[string][]FREDGeoObservation `json:"data"`
}

// FREDGeoObservation represents a single geographic observation
type FREDGeoObservation struct {
	Code      string      `json:"code"`
	Region    string      `json:"region"`
	SeriesID  string      `json:"series_id"`
	Value     interface{} `json:"value"` // Can be string, float64, or int
	Units     string      `json:"units"`
	Frequency string      `json:"frequency"`
}

// FREDGeoRegionalSeries represents parsed regional series data
type FREDGeoRegionalSeries struct {
	Code      string `json:"code"`
	Region    string `json:"region"`
	SeriesID  string `json:"series_id"`
	Value     string `json:"value"`
	Units     string `json:"units"`
	Frequency string `json:"frequency"`
	Date      string `json:"date"`
}
