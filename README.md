# go-user-service
Simple User service for authentication and authorisation of human and system users, written in Go. 

# Endpoints

| Endpoint                | Method  | Description                              |
| ----------------------- | ------- | ---------------------------------------- |
| `/v1/token`             | POST    | Authenticate an existing user and get a token|
| `/v1/user`              | DELETE  | Delete a registered user                 |
| `/v1/user/email/{email}`| GET     | Get a user by their email address        |
| `/v1/user/id/{id}`      | GET     | Get a user by their user ID              |
| `/v1/user`              | POST    | Register a new user                      |
| `/v1/user/activate`     | PUT     | Activate a newly registered user         |
| `/v1/user/authenticate` | POST    | Validate an authentication token         |
