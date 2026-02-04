# go-srv

A Gin server with, 
- a websocket connection that supports concurrent read write between clients
- a webhook endpoint to handle generic payloads 
- a directory event watcher
- a postgres database with routes for secure user management
- jwt authentication 

# docker

<!-- to build -->
docker build -t peterjbishop/go-crypt:latest .
docker run -p 8080:8080 peterjbishop/go-crypt:latest
docker-compose up --build

# ngrok testing

<!-- to test live  -->
ngrok http 8080  
- or launch ngork through docker