# DRYdock

## Features
* Find all docker files recursively down from the current dir
* From n number of containers generate a compose file
  * service for each container with the dir as context
  * Pass env file (that is generated)
  * Add network
  * Add a single service that maps all other services together with "requires"
  * Each dir that contains a Dockerfile file can also compose file that will contain the full service definition. If there is no compose file then default to a simple service def
  *
