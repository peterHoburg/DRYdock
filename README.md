# DRYdock
DRYdock is a project that aims to simplify working locally with many semi-connected projects/microservices.

This was created to help with the following use case:

You work on a team that uses other teams services.
Sometimes you might want to run one or more of those services locally and connect them to your own.
Either for testing, or finding a bug.
To do this with a standard compose file you would need all the services you would ever want to use pulled down, and
if you didn't, docker could throw an error.
With DRYdock all the services you have access to are listed in the UI, and you can pick and choose which ones to run.
DRYdock will then create a compose file with only those services, and run it.


## Installation
* `export DRYDOCK_DIR="<dir on path>" && rm -f $DRYDOCK_DIR/drydock && touch $DRYDOCK_DIR/drydock && wget https://github.com/peterHoburg/DRYdock/releases/latest/download/drydock -O $DRYDOCK_DIR/drydock && chmod +x $DRYDOCK_DIR/drydock`

## Update
Same as installing

## Setup
### Optional
Create a `drydock.yaml` file in the root of the project. This contains keys that will overwrite the default values in the UI

## Usage
* cd into the directory that contains your root docker-compose.yml file
* Run `drydock`
* navigate to localhost:1994 (default port) in your browser
