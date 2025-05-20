package vitalinstrument

import "context"

// Interfaces for reading vital instrument data
type VitalInstrumentGetter interface {
    GetByID(ctx context.Context, id uint) (*VitalInstrument, error)
}

type VitalInstrumentLister interface {
    List(ctx context.Context, filter VitalInstrumentFilter) ([]*VitalInstrument, int, error)
}

type CharacterVitalInstrumentLister interface {
    ListByCharacter(ctx context.Context, characterID uint, limit int) ([]*VitalInstrument, error)
}

// Reader composes all read interfaces
type Reader interface {
    VitalInstrumentGetter
    VitalInstrumentLister
    CharacterVitalInstrumentLister
}
