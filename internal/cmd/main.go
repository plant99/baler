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
	var maxInputFileSize uint64
	var maxInputFileLines uint64
	var maxOutputFileSize uint64
	var exclusionPatterns []string
	var convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert a directory into the minimum number of text files with size < limit specified.",
		Long: `Convert a directory into the minimum number of text files with size < limit specified.

		Arguments: <source-files-directory> <converted-files-directory>

		e.g/

		$ baler convert code_directory/ output_directory/
		`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			convertConfig := &baler.ConvertConfig{
				MaxInputFileLines: maxInputFileLines,
				MaxInputFileSize:  maxInputFileSize,
				MaxOutputFileSize: maxOutputFileSize,
				ExclusionPatterns: &exclusionPatterns,
			}
			err := baler.Convert(args[0], args[1], convertConfig)
			if err != nil {
				// TODO: remove this print
				cmd.PrintErrln(err)
				cmd.Println("Unexpected error occured!")
			}
		},
	}
	convertCmd.Flags().Uint64VarP(&maxInputFileSize, "max-file-size", "u", 1*1024*1024, "Set maximum file size (in bytes) to be considered while converting.")
	convertCmd.Flags().Uint64VarP(&maxInputFileLines, "max-file-lines", "l", 10000, "Set maximum lines a file can have to be considered while converting.")
	convertCmd.Flags().Uint64VarP(&maxOutputFileSize, "max-output-file-size", "s", 5*1024*1024, "Set maximum size (in bytes) of the generated output file.")
	convertCmd.Flags().StringSliceVarP(&exclusionPatterns, "exclusion-patterns", "e", []string{}, "A list of exclusion patterns for baler. e.g '-e \"node_modules*\" -e \"poetry.*\" -e \"package.*\"'")

	// unconvert a group of files into directory
	var unconvertCmd = &cobra.Command{
		Use:   "unconvert",
		Short: "From a group of converted text files, construct the initial directory of files.",
		Long: `Reconstruct the group of text files used for 'baler convert', from the output files of baler.

		Arguments: <converted-files-directory> <source-files-directory>

		e.g/

		$ baler unconvert output_directory/ new_code_directory/
		`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := baler.UnConvert(args[0], args[1])
			if err != nil {
				// TODO: remove this print
				cmd.PrintErrln(err)
				cmd.Println("Unexpected error occured!")
			}
		},
	}
	BalerCommand.AddCommand(versionCmd, convertCmd, unconvertCmd)
}
