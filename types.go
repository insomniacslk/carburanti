package carburanti

import "time"

type Record struct {
	IDImpianto        int
	Carburante        string
	Prezzo            float64
	SelfService       bool
	DataComunicazione time.Time
}

type StationID int

type Station struct {
	ID        StationID
	Gestore   string
	Bandiera  string
	Tipo      StationType
	Nome      string
	Indirizzo string
	Comune    string
	Provincia string
	Lat       string
	Long      string
}

type StationType string

const (
	StationTypeStradale     = "Stradale"
	StationTypeAutostradale = "Autostradale"
)
