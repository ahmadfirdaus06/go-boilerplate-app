## Golang Boilerplate Codebase

### Tech Stack

1. Main Language - Golang https://go.dev
2. Web Server - echo https://echo.labstack.com
3. Database - Local json file 

### Requirements

1. Golang version 1.24+ https://go.dev/doc/install
2. Air (for hot reload in development) https://github.com/air-verse/air

### Features

- [x] Robust http web server
- [x] Easy resource route generator for development (endpoint as CRUD resources)
    - [x] Support nested routing (e.g: /users/:user/apps)
- [x] Adopting proper software development pattern out of the box (route &rarr; controller &rarr; service &rarr; repository &rarr; database)
- [x] Partially ready basic authentication flow
    - [x] Login, Register, Account Verification, JWT authentication

### TODO

- [ ] Support for MongoDB connection
- [ ] More unit/database testing
- [ ] Support for queue dependencies (Redis, Rabbitmq, Apache Kafka)
- [ ] Support robust authentication out of the box
    - [ ] Support for identity provider
    - [ ] Two way authentication
    - [ ] Oauth authentication (Google, Apple ID, Github)
- [ ] Enable queue dependencies for microservice communication
- [ ] Utilize websocket for realtime updates (Websocket, SocketIO)
- [ ] Enable docker support for dockerized development and deployment
- [ ] Implement SDUI support (for better UI maintainability / updates on backend)
    - [ ] Serve html template engine and mounting React components using (Vite + React)

### Setup and Quick Start

1. Clone this repo using SSH ``git clone git@github.com:ahmadfirdaus06/go-boilerplate-app.git``

2. Run `go mod tidy` to install project dependencies

3. Run the project using `air` (if Air already installed) or `go run cmd/http/main.go` instead


### Examples

#### Create simple `notes` CRUD application

1. Refer to file inside `examples/` directory, `/note.route.go`
2. Register this routing inside `/internal/routes/route.go` as follows (within the only function in there at very bottom):

    ``
    examples.InitNoteRoutes(apiRoutes, externals)
    ``
3. Rerun your app, or just let it hot reloading
4. Test the app endpoints to CRUD operations notes (refer the terminal for available methods/routes printed)


### More examples to come...

##### Note: Please be mind this boilerplate is still in early stage and pending for a lot of features and testing. Do not use in production yet unless you have confidence in modifying it. 
    