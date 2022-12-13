package runner

import (
	"strings"

	"github.com/Kelfitas/proxify"
	"github.com/Knetic/govaluate"
	"github.com/projectdiscovery/dsl"
	"github.com/projectdiscovery/gologger"
)

// Runner contains the internal logic of the program
type Runner struct {
	options *Options
	proxy   *proxify.Proxy
}

// NewRunner instance
func NewRunner(options *Options) (*Runner, error) {
	proxy, err := proxify.NewProxy(&proxify.Options{
		Directory:                   options.Directory,
		CertCacheSize:               options.CertCacheSize,
		Verbosity:                   options.Verbosity,
		ListenAddrHTTP:              options.ListenAddrHTTP,
		ListenAddrSocks5:            options.ListenAddrSocks5,
		OutputDirectory:             options.OutputDirectory,
		RequestDSL:                  options.RequestDSL,
		ResponseDSL:                 options.ResponseDSL,
		UpstreamHTTPProxies:         options.UpstreamHTTPProxies,
		UpstreamSock5Proxies:        options.UpstreamSocks5Proxies,
		ListenDNSAddr:               options.ListenDNSAddr,
		DNSMapping:                  options.DNSMapping,
		DNSFallbackResolver:         options.DNSFallbackResolver,
		RequestMatchReplaceDSL:      options.RequestMatchReplaceDSL,
		ResponseMatchReplaceDSL:     options.ResponseMatchReplaceDSL,
		DumpRequest:                 options.DumpRequest,
		DumpResponse:                options.DumpResponse,
		UpstreamProxyRequestsNumber: options.UpstreamProxyRequestsNumber,
		Elastic:                     &options.Elastic,
		Kafka:                       &options.Kafka,
		Allow:                       options.Allow,
		Deny:                        options.Deny,
	})
	if err != nil {
		return nil, err
	}
	return &Runner{options: options, proxy: proxy}, nil
}

// Run polling and notification
func (r *Runner) Run() error {
	if r.options.RequestDSL != "" {
		_, err := govaluate.NewEvaluableExpressionWithFunctions(r.options.RequestDSL, dsl.DefaultHelperFunctions)
		if err != nil {
			printDslCompileError(err)
			return err
		}
	}
	if r.options.ResponseDSL != "" {
		_, err := govaluate.NewEvaluableExpressionWithFunctions(r.options.ResponseDSL, dsl.DefaultHelperFunctions)
		if err != nil {
			printDslCompileError(err)
			return err
		}
	}
	if r.options.RequestMatchReplaceDSL != "" {
		_, err := govaluate.NewEvaluableExpressionWithFunctions(r.options.RequestMatchReplaceDSL, dsl.DefaultHelperFunctions)
		if err != nil {
			printDslCompileError(err)
			return err
		}
	}
	if r.options.ResponseMatchReplaceDSL != "" {
		_, err := govaluate.NewEvaluableExpressionWithFunctions(r.options.ResponseMatchReplaceDSL, dsl.DefaultHelperFunctions)
		if err != nil {
			printDslCompileError(err)
			return err
		}
	}

	// configuration summary
	if r.options.ListenAddrHTTP != "" {
		gologger.Info().Msgf("HTTP Proxy Listening on %s\n", r.options.ListenAddrHTTP)
	}
	if r.options.ListenAddrSocks5 != "" {
		gologger.Info().Msgf("Socks5 Proxy Listening on %s\n", r.options.ListenAddrSocks5)
	}

	if r.options.OutputDirectory != "" {
		gologger.Info().Msgf("Saving proxify traffic to %s\n", r.options.OutputDirectory)
	}
	if r.options.Kafka.Addr != "" {
		gologger.Info().Msgf("Sending traffic to Kafka at %s\n", r.options.Kafka.Addr)
	}
	if r.options.Elastic.Addr != "" {
		gologger.Info().Msgf("Sending traffic to Elasticsearch at %s\n", r.options.Elastic.Addr)
	}

	if len(r.options.UpstreamHTTPProxies) > 0 {
		gologger.Info().Msgf("Using upstream HTTP proxies: %s\n", r.options.UpstreamHTTPProxies)
	} else if len(r.options.UpstreamSocks5Proxies) > 0 {
		gologger.Info().Msgf("Using upstream SOCKS5 proxies: %s\n", r.options.UpstreamSocks5Proxies)
	}

	if r.options.DNSMapping != "" {
		for _, v := range strings.Split(r.options.DNSMapping, ",") {
			gologger.Info().Msgf("Domain => IP: %s\n", v)
		}

		if r.options.DNSFallbackResolver != "" {
			gologger.Info().Msgf("Fallback Resolver: %s\n", r.options.DNSFallbackResolver)
		}

	}

	return r.proxy.Run()
}

// Close the runner instance
func (r *Runner) Close() {
	r.proxy.Stop()
}

// printDslCompileError prints the error message for a DSL compilation error
func printDslCompileError(err error) {
	gologger.Error().Msgf("error compiling DSL: %s", err)
	gologger.Info().Msgf("The available custom DSL functions are:")
	gologger.Info().Label("").Msgf(dsl.GetPrintableDslFunctionSignatures(false))
}
