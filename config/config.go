// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period        time.Duration `config:"period"`
	FetchDuration time.Duration `config:"fetch_duration"`
	FetchTimeout  time.Duration `config:"fetch_timeout"`
	URL           string        `config:"url"`
}

var DefaultConfig = Config{
	Period: 5 * time.Minute,
}
