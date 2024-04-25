package main

import (
	"errors"
	"flag"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/converter/expandconverter"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/nopexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.uber.org/multierr"
	"log"
	"strings"
)

const (
	configFlag = "config"
)

type configFlagValue struct {
	values []string
	sets   []string
}

func (s *configFlagValue) Set(val string) error {
	s.values = append(s.values, val)
	return nil
}

func (s *configFlagValue) String() string {
	return "[" + strings.Join(s.values, ", ") + "]"
}

func flags(reg *featuregate.Registry) *flag.FlagSet {
	flagSet := new(flag.FlagSet)

	cfgs := new(configFlagValue)
	flagSet.Var(cfgs, configFlag, "Locations to the config file(s), note that only a"+
		" single location can be set per flag entry e.g. `--config=file:/path/to/first --config=file:path/to/second`.")

	flagSet.Func("set",
		"Set arbitrary component config property. The component has to be defined in the config file and the flag"+
			" has a higher precedence. Array config properties are overridden and maps are joined. Example --set=processors.batch.timeout=2s",
		func(s string) error {
			idx := strings.Index(s, "=")
			if idx == -1 {
				// No need for more context, see TestSetFlag/invalid_set.
				return errors.New("missing equal sign")
			}
			cfgs.sets = append(cfgs.sets, "yaml:"+strings.TrimSpace(strings.ReplaceAll(s[:idx], ".", "::"))+": "+strings.TrimSpace(s[idx+1:]))
			return nil
		})

	reg.RegisterFlags(flagSet)
	return flagSet
}

func getConfigFlag(flagSet *flag.FlagSet) []string {
	cfv := flagSet.Lookup(configFlag).Value.(*configFlagValue)
	return append(cfv.values, cfv.sets...)
}

func newRootCommand(settings otelcol.CollectorSettings) (root *cobra.Command) {
	flagSet := flags(featuregate.GlobalRegistry())

	root = &cobra.Command{
		Use:          settings.BuildInfo.Command,
		Version:      settings.BuildInfo.Version,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			settings.ConfigProviderSettings.ResolverSettings.URIs = getConfigFlag(flagSet)

			collector, err := otelcol.NewCollector(settings)
			if err != nil {
				return err
			}

			if err = collector.DryRun(cmd.Context()); err != nil {
				// TODO: do the thing I would do here when user config is bad...
				return err
			}

			return collector.Run(cmd.Context())
		},
	}
	root.Flags().AddGoFlagSet(flagSet)
	return
}

func main() {
	settings := otelcol.CollectorSettings{
		BuildInfo: component.BuildInfo{
			Command:     "issue10031",
			Description: "Exhibits problem underlying core issue #10031",
			Version:     "0.1",
		},
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				ProviderFactories: []confmap.ProviderFactory{
					fileprovider.NewFactory(),
				},
				ConverterFactories: []confmap.ConverterFactory{
					expandconverter.NewFactory(),
				},
			},
		},
		Factories: func() (factories otelcol.Factories, errs error) {
			receiverList := []receiver.Factory{
				otlpreceiver.NewFactory(),
			}
			receivers, err := receiver.MakeFactoryMap(receiverList...)
			if err != nil {
				errs = multierr.Append(errs, err)
			}

			exporterList := []exporter.Factory{
				nopexporter.NewFactory(),
			}
			exporters, err := exporter.MakeFactoryMap(exporterList...)
			if err != nil {
				errs = multierr.Append(errs, err)
			}

			extensionList := []extension.Factory{
				healthcheckextension.NewFactory(),
			}
			extensions, err := extension.MakeFactoryMap(extensionList...)
			if err != nil {
				errs = multierr.Append(errs, err)
			}

			factories = otelcol.Factories{
				Receivers:  receivers,
				Exporters:  exporters,
				Extensions: extensions,
			}
			return
		},
	}

	rootCommand := newRootCommand(settings)
	if err := rootCommand.Execute(); err != nil {
		log.Fatalf("collector server run finished with error: %v", err)
	}
}
