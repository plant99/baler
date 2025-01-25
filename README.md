# baler

`baler` is a CLI-tool to convert a text based directory (like a code repository) into a minimal
set of files designed for use with LLMs like Claude, LLama, DeepSeek, ChatGPT.

## Installation

### Binary executables

[GitHub releases of `baler`](https://github.com/plant99/baler/releases) has binary executables for major Operating Systems and Architectures.

### Using `go install`

    $ go install github.com/plant99/baler/cmd/baler@latest

### From source

```sh
git clone https://github.com/plant99/baler.git && cd baler
make build
./bin/baler --help
```

## TL;DR Guide

```
baler convert <source_directory> <destination_directory>
# use/update files in <destination_directory>
baler unconvert <destination_directory> <source_directory>
```

## Intended Usage

`baler`'s intended use is for LLM ingestion.

For interfaces where someone cannot sync their projects automatically like Claude, LLama, *(and unlike Codeium, Copilot)*,
`baler` would help you convert a source code repository into a few text files dealing with which can be much easier.

## Guide

`baler` has 2 major subcommands

1. convert - used to convert a directory into text files.
2. unconvert - used to convert the generated text files back into source directory.

### convert

Example:

    $ baler convert ./ output_dir/ \
      -e LICENSE \
      -e go.sum \
      -e go.mod \
      -e .gitignore \
      -e '.git/*' \
      --delimiter "## filename: " \
      --max-buffer-size 6000000 \
      --verbose \

#### Options

**-d, --delimiter string**

Text that separates 2 files in the generated file.
Note that this delimiter is ALWAYS.
 	- prefixed by a new line ("\n")
 	- suffixed by the next file name and a new line ("\n") (default "// filename: ")

**-e, --exclude strings**

A list of exclusion patterns for baler. e.g '-e "node_modules*" -e "poetry.*" -e "package.*"'

The exclusions should specify the **relative path** to the file/directory that has to be ignored.

Here are some examples
  - to exclude node_modules at root level: `-e node_modules/*`
  - to exclude node_modules inside `frontend`: `frontend/node_modules/*`
  - to exclude a single image inside `static`: `static/logo.png`
  - to exclude the .git directory: `.git/*`

**-b, --max-buffer-size uint**

Set maximum size (in bytes) of buffer for copy operation.

The minimum buffer size reserved is 64kB.
The maximum buffer size defaults to `--max-input-file-size` if this option isn't specified.

**-l, --max-input-file-lines uint**

Set maximum lines a file can have to be considered while converting. (default 10000)

**-i, --max-input-file-size uint**

Set maximum file size (in bytes) to be considered while converting. (default 1048576)

**-o, --max-output-file-size uint**

Set maximum size (in bytes) of the generated output file. (default 5242880)

**-v, --verbose**

Run convert in verbose mode.


#### Notes

1. Double asterisk pattern like `**/node_modules/*` wouldn't work with baler.
2. The minimum "Max Buffer Size" should be equal to the size of the biggest line in your directory. i.e minified CSS, JS files
will require a higher buffer size than human readable source files.


### unconvert

Example:

    $ baler unconvert ./output_dir/ recommended_source/ \
      --max-input-file-size 7000000 \
      --delimiter "## filename: "

#### Options

**-d, --delimiter string**

Text that separates 2 files in the generated file.
This delimiter should be the same one used in the 'convert' command.

**-b, --max-buffer-size uint**

Set maximum size (in bytes) of buffer for copy operation.

The minimum buffer size reserved is 64kB.
The maximum buffer size defaults to `--max-input-file-size` if this option isn't specified.

**-i, --max-input-file-size uint**

Set maximum file size (in bytes) to be considered while converting. (default 1048576)

This should be more than the size of *modified* output of baler convert.

**-v, --verbose**

Run unconvert in verbose mode.

## FAQ / Common Issues

**Q: `baler` stops with an error as soon as it cannot process a file. Shouldn't it continue with other files?**

*Answer*: This is a feature, not a bug ;). This makes an user aware of all the exclusions and edge cases.
There could also be a `--ignore-errors` mode that makes the tool behave this way which will come in if this sees a 0.0.1 release.

**Q: Can I use `baler` in production workflows?**

*Answer*: I wouldn't recommend doing so. At its current state, even the nominal code paths aren't covered fully by unit tests.

Plus, if `baler` errs mid-way it stops but doesn't clean-up its artifacts in the destination directory.
