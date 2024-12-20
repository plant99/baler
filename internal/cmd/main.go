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
	var maxBufferSize uint64
	var exclusionPatterns []string
	var fileDelimiter string
	var verbose bool
	var convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert a directory into smaller text files.",
		Long: `Convert a directory into the minimum number of text files.

		Arguments: <source-files-directory> <converted-files-directory>


		Size Handling:
			- Input files larger than --max-input-size are skipped
			- Output files are split when they reach --max-output-size
			- Read/Write buffer size defaults to input file size if not specified

		e.g/

		$ baler convert code_directory/ output_directory/
		`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			config := &baler.BalerConfig{
				MaxInputFileLines: maxInputFileLines,
				MaxInputFileSize:  maxInputFileSize,
				MaxOutputFileSize: maxOutputFileSize,
				MaxBufferSize:     maxBufferSize,
				ExclusionPatterns: &exclusionPatterns,
				Operation:         baler.OperationConvert,
				FileDelimiter:     fileDelimiter,
				Logger:            newCobraLogger(cmd, verbose),
			}
			// TODO: get output file list
			_, err := baler.Convert(args[0], args[1], config)
			if err != nil {
				handleError(cmd, err)
			}
			// TODO: cmd print success
		},
	}
	convertCmd.Flags().Uint64VarP(&maxInputFileSize, "max-input-file-size", "i", 1*1024*1024, "Set maximum file size (in bytes) to be considered while converting.")
	convertCmd.Flags().Uint64VarP(&maxInputFileLines, "max-input-file-lines", "l", 10000, "Set maximum lines a file can have to be considered while converting.")
	convertCmd.Flags().Uint64VarP(&maxOutputFileSize, "max-output-file-size", "o", 5*1024*1024, "Set maximum size (in bytes) of the generated output file.")
	convertCmd.Flags().Uint64VarP(&maxBufferSize, "max-buffer-size", "b", 0, "Set maximum size (in bytes) of buffer for copy operation.")
	convertCmd.Flags().BoolVarP(&verbose, "verbose", "v", true, "Set log level to verbose.")
	convertCmd.Flags().StringVarP(&fileDelimiter, "delimiter", "d", "filename: ", `Text that separates 2 files in the generated file.

		Note that this delimiter is ALWAYS.
			- prefixed by a new line ("\n")
			- suffixed by the next file name and a new line ("\n")`)
	convertCmd.Flags().StringSliceVarP(&exclusionPatterns, "exclude", "e", []string{}, "A list of exclusion patterns for baler. e.g '-e \"node_modules*\" -e \"poetry.*\" -e \"package.*\"'")

	// unconvert a group of files into directory
	var unconvertCmd = &cobra.Command{
		Use:   "unconvert",
		Short: "Restore original files from converted format.",
		Long: `Reconstruct the group of text files used for 'baler convert', from the output files of baler.

		Arguments: <converted-files-directory> <source-files-directory>

		Buffer size defaults to input file size if not specified.

		e.g/

		$ baler unconvert output_directory/ new_code_directory/
		`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			config := &baler.BalerConfig{
				MaxBufferSize:    maxBufferSize,
				MaxInputFileSize: maxInputFileSize,
				Operation:        baler.OperationUnconvert,
				FileDelimiter:    fileDelimiter,
				Logger:           newCobraLogger(cmd, verbose),
			}
			err := baler.UnConvert(args[0], args[1], config)
			if err != nil {
				handleError(cmd, err)
			}
			// TODO: cmd print success
		},
	}
	unconvertCmd.Flags().Uint64VarP(&maxInputFileSize, "max-input-file-size", "i", 5*1024*1024, "Set maximum size (in bytes) of the input file(s).")
	unconvertCmd.Flags().Uint64VarP(&maxBufferSize, "max-buffer-size", "b", 0, "Set maximum size (in bytes) of buffer for copy operation.")
	unconvertCmd.Flags().BoolVarP(&verbose, "verbose", "v", true, "Set log level to verbose.")
	unconvertCmd.Flags().StringVarP(&fileDelimiter, "delimiter", "d", "// filename: ", `Text that separates 2 files in the generated file.

		Note that this delimiter is ALWAYS.
			- prefixed by a new line ("\n")
			- suffixed by the next file name and a new line ("\n")`)
	BalerCommand.AddCommand(versionCmd, convertCmd, unconvertCmd)
}
