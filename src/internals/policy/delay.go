package policy

import (
	"goblin/src/internals/config"
	"log"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"time"
)

type Delay struct {
	store  *config.ConfigStore
	logger *slog.Logger
}

func NewDelay(store *config.ConfigStore, logger *slog.Logger) *Delay {
	return &Delay{
		store:  store,
		logger: logger,
	}
}

// determine duration based on config - what error was I expecting?
func chooseDelay(cfg config.DelayConfigSettings) (time.Duration, error) {
	if cfg.Mode == config.DelayModeFixed {
		return time.Millisecond * time.Duration(cfg.Min), nil
	}
	return time.Millisecond * time.Duration(rand.Int64N(cfg.Max-cfg.Min)+cfg.Min), nil
}
func (p *Delay) Transport() TransportPolicy {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			config := p.store.Get()

			if !config.DelayConfig.Transport.Enabled {
				log.Println("delay not enabled")
				return next.RoundTrip(req)
			}

			delay, err := chooseDelay(config.DelayConfig.Transport)
			log.Println("delay enabled", "delay", delay)
			if err != nil {
				return nil, err
			}

			timer := time.NewTimer(delay)
			defer timer.Stop()

			select {
			case <-timer.C:
				return next.RoundTrip(req)
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		})
	}
}
func (p *Delay) Response() ResponsePolicy {
	return func(resp *http.Response) error {
		config := p.store.Get()

		if config.Response.ResponseDelay <= 0 {
			return nil
		}

		time.Sleep(time.Duration(config.Response.ResponseDelay))
		return nil
	}
}
