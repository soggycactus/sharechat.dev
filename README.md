# sharechat.dev

A simple chat server that mimics IRC chat rooms:

    - all users are anonymous and assigned a random user name
    - users do not persist after a connection; if you close the room and re-open it in your browser, you will be assigned as a new user
    - messages are only persisted for as long as the room exists

# TODO:

- add ability to delete rooms
- implement short hashing for the room ID
- implement algorithim to generate random username
- add expiration to Rooms
- have Rooms clean themselves up if no members are connected after certain time period