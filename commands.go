package app

import (
	"context"

	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

var subCommands = make([]*cobra.Command, 0)

func SetRootCommand(cmd *cobra.Command) {
	rootCmd = cmd
}

func AddCommands(cmds ...*cobra.Command) {
	subCommands = append(subCommands, cmds...)
}

func executeCommand(appCtx context.Context) error {
	if rootCmd == nil {
		rootCmd = &cobra.Command{
			Use:   "app",                    // TODO changeable
			Short: "This is my application", // TODO changeable
		}
	}

	cfgFile := ""
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "configuration file")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentPreRunE = preRun(&cfgFile, rootCmd.PersistentPreRun, rootCmd.PersistentPreRunE)
	rootCmd.SetContext(appCtx)

	rootCmd.AddCommand(subCommands...)
	return rootCmd.Execute()
}

func preRun(cfgFile *string, runFn func(cmd *cobra.Command, args []string), runErrFn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := initConfig(*cfgFile); err != nil {
			return err
		}

		if config != nil {
			err := initTelemetry(config)
			if err != nil {
				return err
			}
		}

		if runErrFn != nil {
			return runErrFn(cmd, args)
		} else if runFn != nil {
			runFn(cmd, args)
		}
		return nil
	}
}
