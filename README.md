# sharechat.dev

A simple, horizontally scalable chat server that mimics IRC chat rooms:

- all users are anonymous and assigned a random user name
- users do not persist after a connection; if you close the room and re-open it in your browser, you will be registered as a new user
- messages are only persisted for as long as the room exists

## Getting Started

To test the websocket endpoints, I recommend using [websocat](https://github.com/vi/websocat). You can install it via:

    brew install websocat

You can run the full application end-to-end using docker-compose. First spin up Redis (used as a queue) and Postgres (used as the storage):

    docker-compose up -d redis postgres

Once Redis & Postgres are healthy, start the server:
    
    docker-compose up sharechat

### Creating a Room

Rooms are created by sending a POST request to `/api/room`:

    curl -i -X POST localhost:8080/api/room

    HTTP/1.1 200 OK
    Content-Type: application/json
    Date: Fri, 18 Nov 2022 18:15:11 GMT
    Content-Length: 80

    {
        "room_id":"acc3949d-a952-4b0e-8651-13be3cab3563","room_name":"Imperfect Seat"
    }

### Connecting to a room

To connect to a room, connect to the websocket endpoint served at `/api/serve/<room-id>`:

    websocat ws://localhost:8080/api/serve/<room-id>

Once you're connected, you can simply type a message and hit enter to send a message to the room!

![](./docs/websocat.gif)
