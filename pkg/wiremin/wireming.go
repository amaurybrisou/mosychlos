// Package wiremin provides types and utilities for working with the Wiremind API.
package wiremin

type Kind string

const (
	KTimeseries Kind = "ts"
	KQuotes     Kind = "q"
	KNews       Kind = "n"
)

type Payload struct {
	V int     `json:"v"`           // version
	K Kind    `json:"k"`           // kind
	M []any   `json:"m,omitempty"` // meta (array, fixed order)
	D [][]any `json:"d"`           // data rows as arrays
}
