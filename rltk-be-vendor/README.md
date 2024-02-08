# Vendor Service

    This service manages the vandor authentication and manages the vandors actions

## Software Requirement

1. Go - go1.18.3

## API-Docs
    This service exposes REST endpoints to communicate with ZinNext UI. And all the endpoints are properly documented with go-swagger API. It can be used for testing as well.

    The api-doc can be accessed using the `<Base_URL>/api-docs`

## Deployment

    This service is containarised using docker and deployed with other ZinNext services within docker-compose network.

## Run & Test
    Before getting this service up and running, make sure to fill in the configurations in the .env file

    Run - go run main.go

    Test - <Base_URL>/api-docs
