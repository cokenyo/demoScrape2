# demoScrape2

The stats collector for CSC demo files. 

![CSC](https://csconfederation.fandom.com/wiki/Counter_Strike_Confederation_Wiki?file=CSC_Logo.png)

Join the CSC discord -> https://discord.gg/cscc

## Running the parser

Tested on windows.

### Setup 

1. Install [golang](https://go.dev/doc/install).
1. Make a folder named `in` in this directory.
1. Put `.dem` files (not `.zip`) in that folder.

### Run

1. Run with the command `go run .`

Each match will generate a CSV file of stats, a debugging file, and a chatlog in a folder named `out`. 

## Stitching together the outputs

Optionally stitch together the CSV files into one by using the python script.

### Setup 

1. Install [python 3.8](https://www.python.org/downloads/)
1. Install pandas with `pip install pandas`

### Run

1. Run with the command ` python .\stitch_csvs.py`

This will create a file named `monolith.csv` in your `out` folder.


