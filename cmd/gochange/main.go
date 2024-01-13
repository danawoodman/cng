package main

import (
	"fmt"

	"github.com/danawoodman/gochange/internal"
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
		Use:   "gochange [flags] [paths] -- [command]",
		Short: "Runs a command when file changes are detected",
		Run:   execute,
	}
)

func init() {
	// Don't parse commands after the "--" separator
	rootCmd.Flags().SetInterspersed(false)

	rootCmd.PersistentFlags().BoolVarP(&add, "add", "a", false, "Execute command for initially added paths")
	rootCmd.PersistentFlags().BoolVarP(&initial, "initial", "i", false, "Execute command once on load without any event")
	rootCmd.PersistentFlags().StringSliceVarP(&exclude, "exclude", "e", []string{}, "Exclude matching paths")
	rootCmd.PersistentFlags().BoolVarP(&kill, "kill", "k", false, "Kill running processes between changes")
	rootCmd.PersistentFlags().IntVarP(&delay, "delay", "d", 0, "Delay between process changes in milliseconds")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
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
		fmt.Println("No arguments provided\n")
		cmd.Help()
		return
	}

	cmdIndex := indexOf("--", args)
	cmdToRun := args[cmdIndex+1:]

	if len(cmdToRun) == 0 {
		fmt.Println("ERROR: No command specified\n")
		cmd.Help()
		return
	}

	watchedPaths := args[:cmdIndex]

	internal.NewWatcher(&internal.WatcherConfig{
		Command: cmdToRun,
		Paths:   watchedPaths,
		Verbose: verbose,
		Initial: initial,
		Kill:    kill,
		Exclude: exclude,
		Delay:   delay,
	}).Start()
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}
