some python3 CRDT shopping cart


There are client (tcp_serving_client.py) and server (http_server.py) scripts that together form a sample shopping cart app.
Python3.8+ is required to run, dependencies are listed in requirements.txt.

So:
1. Setup postgres database and forward the port to 127.0.0.1:5432 (via docker, ssh port forwarding or whatever)
2. Create database schema by running db_schema.sql against your postgres database
3. Run the server: python3 http_server.py
4. Run the client: python3 tcp_serving_client.py
5. You can send TCP commands to the client to mock user's actions.
Example:
    telnet 127.0.0.1 9999
    add {"product_id":1, "product_name": "test", "price": 1, "quantity": 1}

    telnet 127.0.0.1 9999
    show

    telnet 127.0.0.1 9999
    remove {"product_id":1, "product_name": "test", "price": 1, "quantity": 1}
6. Make some changes on multiple instances of clients and observe the syncronization
between different clients, its (multiple) servers and database.
7. You can observe client state by sending TCP command show to it.
8. Observe server state by `curl -X GET http://127.0.0.1:12347/state`
9. Observe DB state by executing get_db_state.sql against your database.
Note that the DB reports "raw" state, while the client and the server report
current effective state (i.e. hide internal CRDT realization and show the actual cart instead).



Example Docker postgres setup:
docker run -d -p 5432:5432 --name crdt-cart-postgres -e "POSTGRES_USER=server" -e "POSTGRES_PASSWORD=server" postgres:10.7-alpine

# attach to db
docker exec -it crdt-cart-postgres psql -U server

# execute the db_schema.sql script
