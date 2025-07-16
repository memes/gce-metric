package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"github.com/go-logr/zerologr"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appName           = "gce-metric"
	verboseFlagName   = "verbose"
	prettyFlagName    = "pretty"
	projectIDFlagName = "project"
)

var (
	// Version is updated from git tags during build.
	version = "unspecified"
	// ErrFailedToDetectProjectID is returned when the Project ID cannot be determined from Compute metadata.
	ErrFailedToDetectProjectID = errors.New("failed to determine Google project id from operating environment")
)

func newRootCmd() (*cobra.Command, error) {
	cobra.OnInitialize(initConfig)
	rootCmd := &cobra.Command{
		Use:     appName,
		Version: version,
		Short:   "Generate synthetic gauge metrics for Google Cloud Monitoring",
		Long:    `Generate synthetic gauge metrics compatible with Google Cloud Monitoring that follow a cyclic pattern, with values calculated using a range you specify.`,
	}
	rootCmd.PersistentFlags().Count(verboseFlagName, "enable verbose logging; can be repeated to increase verbosity")
	rootCmd.PersistentFlags().Bool(prettyFlagName, false, "disables structured JSON logging to stdout, making it easier to read")
	rootCmd.PersistentFlags().String(projectIDFlagName, "", "the GCP project id to use; specify if not running on GCE or to override detected project id")
	if err := viper.BindPFlag(verboseFlagName, rootCmd.PersistentFlags().Lookup(verboseFlagName)); err != nil {
		return nil, fmt.Errorf("failed to bind '%s' pflag: %w", verboseFlagName, err)
	}
	if err := viper.BindPFlag(prettyFlagName, rootCmd.PersistentFlags().Lookup(prettyFlagName)); err != nil {
		return nil, fmt.Errorf("failed to bind '%s' pflag: %w", prettyFlagName, err)
	}
	if err := viper.BindPFlag(projectIDFlagName, rootCmd.PersistentFlags().Lookup(projectIDFlagName)); err != nil {
		return nil, fmt.Errorf("failed to bind '%s' pflag: %w", projectIDFlagName, err)
	}
	sawtoothCmd := newSawtoothCommand()
	sineCmd := newSineCommand()
	squareCmd := newSquareCommand()
	triangleCmd := newTriangleCommand()
	deleteCmd := newDeleteCommand()
	listCmd, err := newListCommand()
	if err != nil {
		return nil, err
	}
	dataCmd, err := newDataCommand()
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(sawtoothCmd, sineCmd, squareCmd, triangleCmd, deleteCmd, listCmd, dataCmd)
	return rootCmd, nil
}

// Determine the outcome of command line flags, environment variables, and an
// optional configuration file to perform initialization of the application. An
// appropriate zerolog will be assigned as the default logr sink.
func initConfig() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zl := zerolog.New(os.Stderr).With().Caller().Timestamp().Logger()
	viper.AddConfigPath(".")
	if home, err := homedir.Dir(); err == nil {
		viper.AddConfigPath(home)
	}
	viper.SetConfigName("." + appName)
	viper.SetEnvPrefix(appName)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	verbosity := viper.GetInt(verboseFlagName)
	switch {
	case verbosity > 2:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case verbosity == 2:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case verbosity == 1:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
	if viper.GetBool(prettyFlagName) {
		zl = zl.Output(zerolog.ConsoleWriter{
			Out:     os.Stdout,
			NoColor: false,
		})
	}
	logger = zerologr.New(&zl)
	if err == nil {
		return
	}
	var cfgNotFound viper.ConfigFileNotFoundError
	if !errors.As(err, &cfgNotFound) {
		logger.Error(err, "Error reading configuration file")
	}
}

func effectiveProjectID(ctx context.Context) (string, error) {
	logger.V(1).Info("Determining project identifier to use")
	projectID := viper.GetString(projectIDFlagName)
	if projectID != "" {
		logger.V(2).Info("Returning project id from viper", "projectID", projectID)
		return projectID, nil
	}
	if !metadata.OnGCE() {
		return "", ErrFailedToDetectProjectID
	}
	var err error
	if projectID, err = metadata.ProjectIDWithContext(ctx); err != nil {
		return "", fmt.Errorf("failure getting project identifier from metadata: %w", err)
	}
	return projectID, nil
}
