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
	AppName = "gce-metric"
)

var (
	// Version is updated from git tags during build.
	version                    = "unspecified"
	ErrFailedToDetectProjectID = errors.New("failed to determine Google project id from operating environment")
)

func NewRootCmd() (*cobra.Command, error) {
	cobra.OnInitialize(initConfig)
	rootCmd := &cobra.Command{
		Use:     AppName,
		Version: version,
		Short:   "",
		Long:    `Generates synthetic gauge metrics compatible with Google Cloud Monitoring..`,
	}
	rootCmd.PersistentFlags().CountP("verbose", "v", "Enable verbose logging; can be repeated to increase verbosity")
	rootCmd.PersistentFlags().BoolP("pretty", "p", false, "Disables structured JSON logging to stdout, making it easier to read")
	rootCmd.PersistentFlags().String("project", "", "the GCP project id to use; specify if not running on GCE or to override project id")
	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		return nil, fmt.Errorf("failed to bind verbose pflag: %w", err)
	}
	if err := viper.BindPFlag("pretty", rootCmd.PersistentFlags().Lookup("pretty")); err != nil {
		return nil, fmt.Errorf("failed to bind pretty pflag: %w", err)
	}
	if err := viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project")); err != nil {
		return nil, fmt.Errorf("failed to bind project pflag: %w", err)
	}
	sawtoothCmd, err := newSawtoothCommand()
	if err != nil {
		return nil, err
	}
	sineCmd, err := newSineCommand()
	if err != nil {
		return nil, err
	}
	squareCmd, err := newSquareCommand()
	if err != nil {
		return nil, err
	}
	triangleCmd, err := newTriangleCommand()
	if err != nil {
		return nil, err
	}
	deleteCmd, err := newDeleteCommand()
	if err != nil {
		return nil, err
	}
	listCmd, err := newListCommand()
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(sawtoothCmd, sineCmd, squareCmd, triangleCmd, deleteCmd, listCmd)
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
	viper.SetConfigName("." + AppName)
	viper.SetEnvPrefix(AppName)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	verbosity := viper.GetInt("verbose")
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
	if viper.GetBool("pretty") {
		zl = zl.Output(zerolog.ConsoleWriter{Out: os.Stderr})
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
	projectID := viper.GetString("project")
	if projectID != "" {
		logger.V(2).Info("Returning project id from command args", "projectID", projectID)
		return projectID, nil
	}
	if !metadata.OnGCE() {
		return "", ErrFailedToDetectProjectID
	}
	var err error
	if projectID, err = metadata.ProjectID(); err != nil {
		return "", fmt.Errorf("failure getting project identifier from metadata: %w", err)
	}
	return projectID, nil
}
