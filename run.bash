#!/bin/bash

# Check if yq is installed
if ! command -v yq &> /dev/null; then
    echo "yq could not be found. Please install yq to parse YAML files."
    exit 1
fi

# Path to YAML file
YAML_FILE="./config.yaml"

if [ -f "$YAML_FILE" ]; then
    echo "File $YAML_FILE exists. Proceeding to read config."
else
    echo "File $YAML_FILE does not exist. Exiting."
    exit 1
fi

# Extract values from YAML using 
WORKING_DIRECTORY=$(yq '.working_directory' $YAML_FILE)
GO_PROGRAM=$(yq '.go_program' $YAML_FILE)
GO_ARGS=$(yq '.go_args | join(" ")' $YAML_FILE)
CHECK_FILE=$(yq '.check_file' $YAML_FILE)
PYTHON_PROGRAM=$(yq '.python_program' $YAML_FILE)
PYTHON_ARGS=$(yq '.python_args | join(" ")' $YAML_FILE)

# Change directory to the specified working directory
if [ -d "$WORKING_DIRECTORY" ]; then
    echo "Changing directory to $WORKING_DIRECTORY"
    cd "$WORKING_DIRECTORY" || { echo "Failed to change directory to $WORKING_DIRECTORY"; exit 1; }
else
    echo "Working directory $WORKING_DIRECTORY does not exist. Exiting."
    exit 1
fi

# Run Go program
echo "Running Go program..."
go run $GO_PROGRAM $GO_ARGS
GO_EXIT_CODE=$?

# Check if Go program succeeded
if [ $GO_EXIT_CODE -ne 0 ]; then
    echo "Go program failed with exit code $GO_EXIT_CODE"
    exit $GO_EXIT_CODE
fi

# Check if the file exists
if [ -f "$CHECK_FILE" ]; then
    echo "File $CHECK_FILE exists. Proceeding to run Python program."
    
    # Run Python program
    python3 $PYTHON_PROGRAM $PYTHON_ARGS
    PYTHON_EXIT_CODE=$?

    # Check if Python program succeeded
    if [ $PYTHON_EXIT_CODE -ne 0 ]; then
        echo "Python program failed with exit code $PYTHON_EXIT_CODE"
        exit $PYTHON_EXIT_CODE
    fi
else
    echo "File $CHECK_FILE does not exist. Exiting."
    exit 1
fi
