package carburanti

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// See https://www.mimit.gov.it/index.php/it/open-data/elenco-dataset/carburanti-prezzi-praticati-e-anagrafica-degli-impianti
const (
	pricesCSVURL   = "https://www.mimit.gov.it/images/exportCSV/prezzo_alle_8.csv"
	stationsCSVURL = "https://www.mimit.gov.it/images/exportCSV/anagrafica_impianti_attivi.csv"
)

func GetRecords() ([]*Record, error) {
	resp, err := http.Get(pricesCSVURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()
	br := bufio.NewReader(resp.Body)
	// skip the first two lines. This is a non-compliant CSV with a two-line
	// header.
	for i := 0; i < 2; i++ {
		if _, _, err := br.ReadLine(); err != nil {
			return nil, fmt.Errorf("failed to read line: %w", err)
		}
	}
	r := csv.NewReader(br)
	r.Comma = ';'
	r.FieldsPerRecord = 5
	var records []*Record
	for {
		items, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %w", err)
		}
		record, err := parseRecord(items)
		if err != nil {
			return nil, fmt.Errorf("failed to parse record: %w", err)
		}
		records = append(records, record)
	}
	return records, nil
}

func parseRecord(items []string) (*Record, error) {
	if len(items) != 5 {
		return nil, fmt.Errorf("expected 5 fields, got %d", len(items))
	}
	var r Record

	idImpianto, err := strconv.ParseInt(items[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("IDImpianto is not a numeric string: %w", err)
	}
	r.IDImpianto = int(idImpianto)
	r.Carburante = items[1]
	r.Prezzo, err = strconv.ParseFloat(items[2], 64)
	if err != nil {
		return nil, fmt.Errorf("Prezzo is not a float string: %w", err)
	}
	r.SelfService, err = strconv.ParseBool(items[3])
	if err != nil {
		return nil, fmt.Errorf("SelfService is not a bool string: %w", err)
	}
	r.DataComunicazione, err = time.Parse("2/1/2006 15:04:05", items[4])
	if err != nil {
		return nil, fmt.Errorf("DataComunicazione is not a time string: %w", err)
	}

	return &r, nil
}

func GetStations() (map[StationID]Station, error) {
	resp, err := http.Get(stationsCSVURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch station data: %w", err)
	}
	defer resp.Body.Close()
	br := bufio.NewReader(resp.Body)
	// skip the first two lines. This is a non-compliant CSV with a two-line
	// header.
	for i := 0; i < 2; i++ {
		if _, _, err := br.ReadLine(); err != nil {
			return nil, fmt.Errorf("failed to read line: %w", err)
		}
	}
	r := csv.NewReader(br)
	r.Comma = ';'
	r.FieldsPerRecord = 10
	r.LazyQuotes = true
	stationMap := make(map[StationID]Station)
	for {
		items, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if errors.Is(err, &csv.ParseError{}) {
				if strings.Contains(err.Error(), "wrong number of fields") {
					log.Printf("Skipping malformed record with wrong number of fields")
					continue
				}
			}
		}
		idImpianto, err := strconv.ParseInt(items[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("IDImpianto is not a numeric string: %w", err)
		}
		_, ok := stationMap[StationID(idImpianto)]
		if ok {
			log.Printf("Found duplicate type '%s' for station ID %d", items[3], idImpianto)
		}
		stationMap[StationID(idImpianto)] = Station{
			ID:        StationID(idImpianto),
			Gestore:   items[1],
			Bandiera:  items[2],
			Tipo:      StationType(items[3]),
			Nome:      items[4],
			Indirizzo: items[5],
			Comune:    items[6],
			Provincia: items[7],
			Lat:       items[8],
			Long:      items[9],
		}
	}
	return stationMap, nil
}
