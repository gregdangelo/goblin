package policy

import (
	"errors"
	"fmt"
	"goblin/src/internals/config"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	return f(req)
}

type TransportPolicy func(http.RoundTripper) http.RoundTripper

func ChainTransport(base http.RoundTripper, policies ...TransportPolicy) http.RoundTripper {
	for i := len(policies) - 1; i >= 0; i-- {
		base = policies[i](base)
	}
	return base
}

type ResponsePolicy func(*http.Response) error

func ChainResponse(policies ...ResponsePolicy) ResponsePolicy {
	return func(resp *http.Response) error {
		for _, policy := range policies {
			if err := policy(resp); err != nil {
				return err
			}
		}
		return nil
	}
}

type TransportPolicyRule interface {
	Transport() TransportPolicy
}
type ResponsePolicyRule interface {
	Response() ResponsePolicy
}

type Policies struct {
	Transport []TransportPolicy
	Response  []ResponsePolicy
}

func LoadPolicies(store *config.ConfigStore, logger *slog.Logger) Policies {
	policies := Policies{}
	var allPolicies interface{} = NewDelay(store, logger)
	if transport, ok := allPolicies.(TransportPolicyRule); ok {
		policies.Transport = append(policies.Transport, transport.Transport())
	}
	if resp, ok := allPolicies.(ResponsePolicyRule); ok {
		policies.Response = append(policies.Response, resp.Response())
	}
	return policies
}

// these need to be updated and loaded in
func FailurePolicy(store *config.ConfigStore) TransportPolicy {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			config := store.Get()

			if config.FailRequestsConfig.Enabled {
				return nil, errors.New("injected connection failure")
			}

			return next.RoundTrip(req)
		})
	}
}

func InjectStatus(code int, body string) func(*http.Response) error {
	return func(resp *http.Response) error {
		resp.StatusCode = code
		resp.Status = fmt.Sprintf("%d %s", code, http.StatusText(code))
		resp.Body = io.NopCloser(strings.NewReader(body))
		resp.ContentLength = int64(len(body))
		resp.Header.Set("Content-Type", "text/plain")
		resp.Header.Set("Content-Length", strconv.Itoa(len(body)))

		return nil
	}
}
