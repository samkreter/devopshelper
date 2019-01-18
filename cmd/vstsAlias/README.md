# VSTS Alias Converter

Convert a list of aliases to the needed information for the vsts auto reviewer

## Getting Started

### Building

From the root directory of the vsts auto reviewer run the following command to build the container

    docker build -t vstsalias:0.0.1 -f ./cmd/vstsAlias/Dockerfile .

### Usage

You must set the CONFIG_PATH enviorment variable to the same file format as the vsts autoreviewer

Convert a list of aliases to their IDs

    docker run -e CONFIG_PATH="/path/to/config.json" vstsalias:0.0.1 -a myalias,anotheralias

Convert a list of aliases in a text file

    docker run -e CONFIG_PATH="/path/to/config.json" vstsalias:0.0.1 -f aliases.txt

Convert a list of aliases to json format for the autoreviewer

    docker run -e CONFIG_PATH="/path/to/config.json" vstsalias:0.0.1 -f aliases.txt -o reviewers.json