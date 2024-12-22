package cmd

import (
	"fmt"

	"github.com/plant99/baler/internal/baler"
	"github.com/spf13/cobra"
)

var BalerCommand = &cobra.Command{
	Use:   "baler",
	Short: "Convert text directories into minimum number of files to use with LLMs.",
	Long:  `baler converts, unconverts a text directory (e.g/ a code repository) into the minimum number of files such that each file is smaller than a given size limit.`,
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
	var convertMaxInputFileSize, unconvertMaxInputFileSize uint64
	var maxInputFileLines uint64
	var maxOutputFileSize uint64
	var convertMaxBufferSize, unconvertMaxBufferSize uint64
	var exclusionPatterns []string
	var convertFileDelimiter, unconvertFileDelimiter string
	var convertVerbose, unconvertVerbose bool
	var convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert a directory into smaller text files.",
		Long: `Convert a directory into the minimum number of text files.

Arguments: <source-files-directory> <converted-files-directory>


Size Handling:
	- Input files larger than --max-input-file-size are skipped
	- Output files are split when they reach --max-output-file-size
	- Read/Write buffer size defaults to "--max-input-file-size" if not specified

e.g/

$ baler convert code_directory/ output_directory/
		`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			config := &baler.BalerConfig{
				MaxInputFileLines: maxInputFileLines,
				MaxInputFileSize:  convertMaxInputFileSize,
				MaxOutputFileSize: maxOutputFileSize,
				MaxBufferSize:     convertMaxBufferSize,
				ExclusionPatterns: &exclusionPatterns,
				Operation:         baler.OperationConvert,
				FileDelimiter:     convertFileDelimiter,
				Logger:            newCobraLogger(cmd, convertVerbose),
				Verbose:           convertVerbose,
			}
			// validation
			if config.MaxInputFileSize >= config.MaxOutputFileSize {
				handleError(
					cmd,
					fmt.Errorf(
						"--max-input-file-size = %d cannot be > --max-output-file-size = %d",
						config.MaxInputFileSize,
						config.MaxOutputFileSize,
					),
				)
			}
			// TODO: get output file list
			_, err := baler.Convert(args[0], args[1], config)
			if err != nil {
				handleError(cmd, err)
			}
			cmd.Println("Conversion successful!")
		},
	}
	convertCmd.Flags().Uint64VarP(&convertMaxInputFileSize, "max-input-file-size", "i", 1*1024*1024, "Set maximum file size (in bytes) to be considered while converting.")
	convertCmd.Flags().Uint64VarP(&maxInputFileLines, "max-input-file-lines", "l", 10000, "Set maximum lines a file can have to be considered while converting.")
	convertCmd.Flags().Uint64VarP(&maxOutputFileSize, "max-output-file-size", "o", 5*1024*1024, "Set maximum size (in bytes) of the generated output file.")
	convertCmd.Flags().Uint64VarP(&convertMaxBufferSize, "max-buffer-size", "b", 0, "Set maximum size (in bytes) of buffer for copy operation.")
	convertCmd.Flags().BoolVarP(&convertVerbose, "verbose", "v", false, "Run baler in verbose mode.")
	convertCmd.Flags().StringVarP(
		&convertFileDelimiter,
		"delimiter",
		"d",
		"// filename: ",
		`Text that separates 2 files in the generated file.
Note that this delimiter is ALWAYS.
	- prefixed by a new line ("\n")
	- suffixed by the next file name and a new line ("\n")`,
	)
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
				MaxBufferSize:    unconvertMaxBufferSize,
				MaxInputFileSize: unconvertMaxInputFileSize,
				Operation:        baler.OperationUnconvert,
				FileDelimiter:    unconvertFileDelimiter,
				Logger:           newCobraLogger(cmd, unconvertVerbose),
				Verbose:          unconvertVerbose,
			}
			err := baler.UnConvert(args[0], args[1], config)
			if err != nil {
				handleError(cmd, err)
			}
			cmd.Println("Un-conversion successful!")
		},
	}
	unconvertCmd.Flags().Uint64VarP(&unconvertMaxInputFileSize, "max-input-file-size", "i", 5*1024*1024, "Set maximum size (in bytes) of the input file(s).")
	unconvertCmd.Flags().Uint64VarP(&unconvertMaxBufferSize, "max-buffer-size", "b", 0, "Set maximum size (in bytes) of buffer for copy operation.")
	unconvertCmd.Flags().BoolVarP(&unconvertVerbose, "verbose", "v", false, "Run baler in verbose mode.")
	unconvertCmd.Flags().StringVarP(
		&unconvertFileDelimiter,
		"delimiter",
		"d",
		"// filename: ",
		`Text that separates 2 files in the generated file.
Note that this delimiter is ALWAYS.
	- prefixed by a new line ("\n")
	- suffixed by the next file name and a new line ("\n")`,
	)
	BalerCommand.AddCommand(versionCmd, convertCmd, unconvertCmd)
}
