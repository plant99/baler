# baler

`baler` is a tool to convert a text based directory (like a code repository) into a minimal
set of files designed for use with LLMs like Claude, LLama.

## Installation

### Binary Executables

### Using `go install`

    $ go install github.com/plant99/baler/cmd/baler@latest

### From Source

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
For interfaces where someone cannot sync their projects like Claude, LLama, *(and unlike Codeium, Copilot)*,
`baler` would help you convert a source code repository into a few text files dealing with which can be much easier.

It comes with some useful features like

  - **File Size Limits**: The generated text files can be limited to a particular size. The default for baler is 5 MB, which can be controlled by the user.
  - **Preserving Structure**: The generated text files hints the LLM  about the relative file path before presenting its content.
  In fact, one can ask the LLM to preserve this file structure and then baler can help you `unconvert` the edited file into your source directory.
  - **Other Validations**: There are other validations in place like placing a limit on line count (to prevent lock files from being considered), UTF-8 validity checks(to prevent non-text formats).
  - **Exclusion Patterns**: `baler` allows you to skip directories like `.git`, `node_modules/*`  with a `-e` flag.

## Guide

## FAQ

## Miscellaneous

### Where does the name come from?
