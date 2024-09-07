# Open Dependency Analysis 📜

This program was created to help me analyze programs written in Python. I want to have an easy way to read a file, 
create a dependency matrix, and have a graphical overview of how the program looks and how the files interact.

Other programs exist, such as [CodeScene](https://codescene.io), but I had to create my own for reasons I am unwilling to share.

## Functionallity

The program has three main parts: the parser (written in Go), the analysis tool (written in Python), and the outputs.
The initial release's outputs include the option to export a CSV file to be used as an initial step in creating a dependency matrix.

The second output option is an interactive graph created with [Pyvis](https://pyvis.readthedocs.io/en/latest/). Upon program completion, it will open in the default web browser.
If it doesn't open itself, one can manually open the file `src/da_network_view.html` to achieve the same result.

## Instructions

### Build

TODO

### Run

Create a `config.yaml` file in the repository directory's base. Use the following template.

```yaml

working_directory: "./src"

go_program: "main.go"
go_args: ["", "python"]

check_file: "data.json"

python_program: "dependency_analyzer.py"
python_args: ["", ""]

```

Run the program with the included bash script.

```bash
./run.bash
```

### Available arguments

- **go_args**
  - First argument: Path to the directory to analyze
  - Second argument: Programming language. Currently, only "python" is available.
- **python_args**
```
usage: dependency_analyzer.py [-h] [--no_cache] [--create_csv] [--create_network_view] [--dynamic_node_size] [--dynamic_edge_size]
                              [-l LIMIT]

options:
  -h, --help            show this help message and exit
  --no_cache            forces the usage of no-cached values
  --create_csv          create a CSV file to use as a dependency matrix
  --create_network_view
                        include to create an interactive network view of the dependencies
  --dynamic_node_size   include to enable the usage of dynamic node sizes in the network view
  --dynamic_edge_size   include to enable the usage of dynamic edge sizes in the network view
  -l LIMIT, --limit LIMIT
                        set a limit for the count of references to a given package/module to be included
```
