package cmd

import (
	"github.com/plant99/baler/internal/baler"
	"github.com/spf13/cobra"
)

var BalerCommand = &cobra.Command{
	Use:   "baler",
	Short: "Convert text directories into minimum number of files to use with LLMs.",
	Long: `baler converts, unconverts a text directory (e.g/ a code repository) into the minimum number of files such that
	each file is smaller than a given size limit.`,
}

func Run() {
	AddCommands()
	BalerCommand.Execute()
}

func AddCommands() {
	// version command
	version := "0.0.1-beta"
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the version of baler",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("baler: ", version)
		},
	}

	// convert a repository
	var convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert a directory into the minimum number of text files with size < limit specified.",
		Long: `Convert a directory into the minimum number of text files with size < limit specified.

		Arguments: <path>
		`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := baler.Convert(args[0], uint32(5120))
			if err != nil {
				cmd.Println("Unexpected error occured!")
			}
		},
	}
	BalerCommand.AddCommand(versionCmd, convertCmd)
}
