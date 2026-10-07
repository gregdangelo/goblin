package controller

import (
	"errors"
	"goblin/src/internals/config"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

var (
	errOutofBoundsMS = errors.New("ms is out of acceptable bounds")
)

/*
I don't need partials I can save it all as one
*/
type WebHandler interface {
	Http(w http.ResponseWriter, r *http.Request)
}

type AdminHandler struct {
	store    *config.ConfigStore
	renderer *Renderer
	path     string
	logger   *slog.Logger
}

func New(store *config.ConfigStore, r *Renderer, logger *slog.Logger) WebHandler {
	return &AdminHandler{
		store:    store,
		renderer: r,
		logger:   logger,
		path:     "pages/",
	}
}

func enabled(s string) bool {
	return s != "" && s == "on"
}
func checked(b bool) string {
	if !b {
		return ""
	}
	return "checked"
}

func strToIntConfig(s string) (int64, error) {
	str := strings.TrimSpace(s)
	if str == "" {
		return 0, nil
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	// boundary checking - less than 1 minute greater or equal than 0
	if i < 0 || i > 60000 {
		return 0, errOutofBoundsMS
	}

	return int64(i), nil
}

// this hopefully will work but I've adjusted the config structure since
func validateDelay(r *http.Request) (*config.DelayConfig, error) {
	cfg := config.DelayConfig{
		Transport: config.DelayConfigSettings{
			Enabled: r.FormValue("delay_enable") != "",
		},
	}

	minVal, err := strToIntConfig(r.FormValue("delay_min"))
	if err != nil {
		return nil, err
	}
	cfg.Transport.Min = minVal

	maxVal, err := strToIntConfig(r.FormValue("delay_max"))
	if err != nil {
		return nil, err
	}
	cfg.Transport.Max = maxVal

	return &cfg, nil
}

func (h *AdminHandler) Http(w http.ResponseWriter, r *http.Request) {
	// We need a super simple admin panel
	if r.Header.Get("hx-request") == "true" {
		var errs []error
		enabledField := r.FormValue("delay_enable") != ""
		delayEnable := enabled(r.FormValue("delay_enable"))
		failureEnable := enabled(r.FormValue("failure_enable"))
		h.logger.Info("values",
			"enabledField", enabledField,
			"delayEnable", delayEnable,
			"delayMax", 0,
			"failureEnable", failureEnable,
		)
		defaultConfig := h.store.Get()
		cfg := config.ChaosConfig{}
		dly, err := validateDelay(r)
		if err != nil {
			/*
				add the error and overwrite with the default
				honestly we should preserve the good values but
				that can be a future iteration.  For now just
				don't mess up.  Plus it only overwrites that
				single section where the error actually is
			*/
			errs = append(errs, err)
			dly = &defaultConfig.DelayConfig
		}
		cfg.DelayConfig = *dly

		// idealy we turn Config into this data
		data := map[string]any{
			"DelayMax":       cfg.DelayConfig.Response.Max,
			"DelayMin":       cfg.DelayConfig.Response.Min,
			"DelayEnabled":   checked(cfg.DelayConfig.Response.Enabled),
			"FailureEnabled": checked(failureEnable),
		}

		if len(errs) == 0 {
			h.store.Set(cfg)
			// Overwrite the existing configuration file
			if r.FormValue("enable_overwrite") != "" {
				h.store.Save("goblin.toml")
			}
		} else {
			errStr := make([]string, len(errs))
			for i, e := range errs {
				errStr[i] = e.Error()
			}
			data["Error"] = strings.Join(errStr, "<br/>\n")
		}
		h.RenderComponent(w, r, "home", data)
		return
	}
	cfg := h.store.Get()
	data := map[string]any{
		"DelayMax":       cfg.DelayConfig.Response.Max,
		"DelayMin":       cfg.DelayConfig.Response.Min,
		"DelayEnabled":   checked(cfg.DelayConfig.Response.Enabled),
		"FailureEnabled": checked(cfg.FailRequestsConfig.Enabled),
	}
	h.RenderPage(w, r, "home", data)

}

// from a previous project where there were multiple pages but is less useful here
func (h *AdminHandler) RenderPage(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	if err := h.renderer.Render(w, h.path+name, "base.html", data); err != nil {
		h.logger.Error("failed to render page", "error", err)
	}
}
func (h *AdminHandler) RenderComponent(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	if err := h.renderer.Render(w, h.path+name, "content", data); err != nil {
		h.logger.Error("failed to render component", "error", err)
	}
}

func MatchNoopPaths(s string) bool {
	knownPaths := []string{
		"/.well-known/appspecific/com.chrome.devtools.json",
		"/favicon.ico",
	}
	for _, path := range knownPaths {
		if strings.HasPrefix(s, path) {
			return true
		}
	}
	return false
}
