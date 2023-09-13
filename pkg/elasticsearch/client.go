package elastic8

import (
	"emperror.dev/errors"
	"github.com/cenkalti/backoff/v4"
	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
	configutil "github.com/je4/utils/v2/pkg/config"
	"github.com/rs/zerolog"
	"os"
	"time"
)

func NewClient(address string, index string, apikey string, logger *zerolog.Logger) (*elasticsearch.TypedClient, error) {
	cfg := &Elastic8Config{
		Adresses:               []string{address},
		Username:               "",
		Password:               "",
		CACert:                 "",
		CertificateFingerprint: "",
		ServiceToken:           configutil.EnvString(apikey),
	}
	return NewClient2(index, cfg, logger)
}

func NewClient2(index string, cfg *Elastic8Config, logger *zerolog.Logger) (*elasticsearch.TypedClient, error) {

	var err error
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	//
	// Use a third-party package for implementing the backoff function
	//
	retryBackoff := backoff.NewExponentialBackOff()

	var certBytes []byte
	if cfg.CACert != "" {
		certBytes, err = os.ReadFile(cfg.CACert)
		if err != nil {
			//logger.Err(err).Msgf("cannot read cert file '%s'", cfg.CACert)
			return nil, errors.Wrapf(err, "cannot read cert file '%s'", cfg.CACert)
		}
	}

	elasticConfig := elasticsearch.Config{
		APIKey:                 string(cfg.APIKey),
		Addresses:              cfg.Adresses,
		CloudID:                cfg.CloudID,
		ServiceToken:           string(cfg.ServiceToken),
		Username:               string(cfg.Username),
		Password:               string(cfg.Password),
		CertificateFingerprint: string(cfg.CertificateFingerprint),
		CACert:                 certBytes,

		// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
		// ... using the "apmelasticsearch" wrapper for instrumentation
		//		Transport: apmelasticsearch.WrapRoundTripper(http.DefaultTransport),
		// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

		// Retry on 429 TooManyRequests statuses
		//
		RetryOnStatus: []int{502, 503, 504, 429},
		// Configure the backoff function
		//
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},

		// Retry up to 5 attempts
		//
		MaxRetries: 5,

		Logger: &elastictransport.ColorLogger{Output: os.Stdout},
	}

	tc, err := elasticsearch.NewTypedClient(elasticConfig)
	return tc, errors.Wrapf(err, "cannot create client")
}
