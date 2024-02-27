package main

import (
	"fmt"
	"os"

	"github.com/danawoodman/cng/internal"
	"github.com/spf13/cobra"
)

var (
	add       bool
	initial   bool
	exclude   []string
	noExclude bool
	kill      bool
	jobs      int
	delay     int
	await     int
	poll      int
	outpipe   string
	filter    string
	verbose   bool

	rootCmd = &cobra.Command{
		Use:   "cng [flags] [paths] -- [command]",
		Short: "Runs a command when file changes are detected",
		Run:   execute,
	}
)

func init() {
	// Don't parse commands after the "--" separator
	rootCmd.Flags().SetInterspersed(false)

	rootCmd.PersistentFlags().BoolVarP(&add, "add", "a", false, "execute command for initially added paths")
	rootCmd.PersistentFlags().BoolVarP(&initial, "initial", "i", false, "execute command once on load without any event")
	rootCmd.PersistentFlags().StringSliceVarP(&exclude, "exclude", "e", []string{}, "exclude matching paths")
	rootCmd.PersistentFlags().BoolVarP(&kill, "kill", "k", false, "kill running processes between changes")
	rootCmd.PersistentFlags().IntVarP(&delay, "delay", "d", 0, "delay between process changes in milliseconds")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	// TODO: implement more onchange features:
	// rootCmd.PersistentFlags().BoolVar(&noExclude, "no-exclude", false, "Disable default exclusion")
	// rootCmd.PersistentFlags().IntVarP(&jobs, "jobs", "j", 1, "Set max concurrent processes")
	// rootCmd.PersistentFlags().IntVar(&await, "await-write-finish", 2000, "Hold events until the size doesn't change")
	// rootCmd.PersistentFlags().IntVarP(&poll, "poll", "p", 0, "Use polling for change detection")
	// rootCmd.PersistentFlags().StringVarP(&outpipe, "outpipe", "o", "", "Shell command to execute on every change")
	// rootCmd.PersistentFlags().StringVarP(&filter, "filter", "f", "", "Filter events to listen")
}

func main() {
	rootCmd.Execute()
}

func execute(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fail(cmd, "No arguments provided, please at least pass a pattern to watch and a command to run")
	}

	cmdIndex := indexOf("--", args)
	if cmdIndex == -1 {
		fail(cmd, "No '--' separator found between paths and command")
	}
	cmdToRun := args[cmdIndex+1:]

	if len(cmdToRun) == 0 {
		fail(cmd, "No command specified, pass a command to run after the '--' separator")
	}

	watchedPaths := args[:cmdIndex]

	internal.NewWatcher(internal.WatcherConfig{
		Command:      cmdToRun,
		ExcludePaths: watchedPaths,
		Verbose:      verbose,
		Initial:      initial,
		Kill:         kill,
		Exclude:      exclude,
		Delay:        delay,
	}).Start()
}

func fail(cmd *cobra.Command, msg string) {
	fmt.Println(msg)
	fmt.Println("")
	cmd.Help()
	os.Exit(1)
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}
