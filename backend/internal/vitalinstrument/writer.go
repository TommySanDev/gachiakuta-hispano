package vitalinstrument

import "context"

// Interfaces for writing vital instrument data
type VitalInstrumentCreator interface {
    Create(ctx context.Context, instrument *VitalInstrument) error
}

type VitalInstrumentUpdater interface {
    Update(ctx context.Context, instrument *VitalInstrument) error
}

type VitalInstrumentDeleter interface {
    Delete(ctx context.Context, id uint) error
}

type VitalInstrumentRestorer interface {
    Restore(ctx context.Context, id uint) error
}

type VitalInstrumentPermanentDeleter interface {
    DeletePermanently(ctx context.Context, id uint) error
}

// Writer composes all write interfaces
type Writer interface {
    VitalInstrumentCreator
    VitalInstrumentUpdater
    VitalInstrumentDeleter
    VitalInstrumentRestorer
    VitalInstrumentPermanentDeleter
}

// Store composes all read and write operations
type Store interface {
    Reader
    Writer
}
