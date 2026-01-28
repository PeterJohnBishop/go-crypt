# go-crypt

# websocket package

# hub
- there is a single instance of the hub that maintains a list of connected clients, and contains channels that register and deregister clients, and transfer data.
- when a client app connects to the websocket a new websocket client is created and registers with the hub
- the hub then waits for data

# client
- when a client app sends data to the websocket, the client readPump sends that data to hub.broadcast where's its processed - when the hub recieves data in the hub.broadcast channel it loops throgh the clients list and sends the message to each client's send channel
- when data arrives in the send channel the writePump sends that data to the websocket and off to the client app

# totp
- on websocket connect the client app generates a random id and random secret and sends these to the hub through request headers
- client id and client secret are stored locally
- the websocket client generates a totp uri and sends this via email as a PNG
- the client id and client secret are stored in the hub 