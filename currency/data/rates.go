package data

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hashicorp/go-hclog"
)

type ExchangeRates struct {
	l     hclog.Logger
	rates map[string]float64
}

func NewRates(l hclog.Logger) (*ExchangeRates, error) {
	er := &ExchangeRates{l: l, rates: map[string]float64{}}

	err := er.getRates()

	return er, err
}

func (e *ExchangeRates) GetRate(base string, destination string) (float64, error) {
	br, ok := e.rates[base]
	if !ok {
		return 0, fmt.Errorf("Rate not found for currency %s", base)
	}

	dr, ok := e.rates[destination]
	if !ok {
		return 0, fmt.Errorf("Rate not found for currency %s", destination)
	}

	return dr / br, nil
}

func (e *ExchangeRates) MonitorRates(interval time.Duration) chan struct{} {
	ret := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				for k, v := range e.rates {
					change := (rand.Float64() / 10)
					direction := rand.Intn(1)

					if direction == 0 {
						change = 1 - change
					} else {
						change = 1 + change
					}

					e.rates[k] = v * change
				}

				ret <- struct{}{}
			}
		}
	}()

	return ret
}

func (e *ExchangeRates) getRates() error {
	// resp, err := http.DefaultClient.Get("https://www.ecb.europa.eu/stats/eurofyref/eurofyref-daily.xml")
	// if err != nil {
	// 	return nil
	// }

	// if resp.StatusCode != http.StatusOK {
	// 	return fmt.Errorf("Expected status code 200 but got %d", resp.StatusCode)
	// }
	// defer resp.Body.Close()

	// md := &Cubes{}
	// xml.NewDecoder(resp.Body).Decode(&md)

	for _, c := range []string{"GBP", "USD", "JPY", "BGN"} {
		// r, err := strconv.ParseFloat(c, 64)
		// if err != nil {
		// 	return err
		// }

		e.rates[c] = (rand.Float64() / 3)
	}

	// Data gives rates in comparison to eur
	e.rates["EUR"] = 1

	return nil
}

type Cubes struct {
	CubeData []Cube `xml:"Cube>Cube>Cube"`
}

type Cube struct {
	Currency string `xml:"currency,attr"`
	Rate     string `xml:"rate,attr"`
}
