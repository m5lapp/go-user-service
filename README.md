# go-user-service
Simple User service for authentication and authorisation of human and system users, written in Go. 

# Endpoints

| ? | Endpoint           | Method  | Description                               |
| - | -------------------| ------  | ----------------------------------------- |
| Y | `/v1/token`        | POST    | Authenticate an existing user and get a token|
| Y | `/v1/user`         | DELETE  | Delete a registered user                  |
| Y | `/v1/user`         | POST    | Register a new user                       |
| Y | `/v1/user/activate`| PUT     | Activate a newly registered user          |
| Y | `/v1/user/authenticate` | POST | Validate an authentication token        |
