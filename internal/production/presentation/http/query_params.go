package http

import (
	"net/url"
	"strconv"
)

func parseOptionalInt(values url.Values, key string) (*int, error) {
	if !values.Has(key) { return nil, nil }
	value, err := strconv.Atoi(values.Get(key))
	if err != nil { return nil, err }
	return &value, nil
}

func optionalString(values url.Values, key string) *string {
	if !values.Has(key) { return nil }
	value := values.Get(key)
	return &value
}
