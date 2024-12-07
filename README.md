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
* `curl -sSL https://raw.githubusercontent.com/peterHoburg/DRYdock/refs/heads/main/INSTALL | bash`
## Update
Same as installing

## Setup
### Optional
Create a `drydock.yaml` file in the root of the project. This contains keys that will overwrite the default values in the UI

## Usage
* cd into the directory that contains your root docker-compose.yml file
* Run `drydock`
* navigate to localhost:1994 (default port) in your browser
