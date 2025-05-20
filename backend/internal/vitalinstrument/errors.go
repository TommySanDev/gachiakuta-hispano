package vitalinstrument

import "errors"

var (
    ErrVitalInstrumentNotFound = errors.New("vital instrument not found")
    ErrInvalidInput = errors.New("invalid input")
    ErrVitalInstrumentAlreadyExists = errors.New("vital instrument already exists")
)
