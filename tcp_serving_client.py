import json
import os
import socket
import threading
import time
import socketserver

from crdt_cart.client import ClientCartSyncronizer
from crdt_cart.common.cart import CartItem


SERVER_SYNC_URL = 'http://%s:12347/sync_state' % os.environ['SERVER_HOST']

CLIENT_STATE = ClientCartSyncronizer(SERVER_SYNC_URL)


class MyTCPHandler(socketserver.BaseRequestHandler):
    """
    The request handler class for our server.

    It is instantiated once per connection to the server, and must
    override the handle() method to implement communication to the
    client.
    """

    def handle(self):
        # self.request is the TCP socket connected to the client
        data = self.request.recv(1024).strip()
        print("{} wrote:".format(self.client_address[0]))
        print(data)
        if data.startswith(b'add'):
            to_add = json.loads(data[len('add'):].strip())
            elem = CartItem(**to_add)
            print('adiing %s to cart' % to_add)
            CLIENT_STATE.current_cart.add(elem)
            CLIENT_STATE.sync_with_server()
            self.request.sendall(b'success\n')
        elif data.startswith(b'remove'):
            to_remove = json.loads(data[len('remove'):].strip())
            elem = CartItem(**to_remove)
            print('removing %s from cart' % to_remove)
            CLIENT_STATE.current_cart.remove(elem)
            CLIENT_STATE.sync_with_server()
            self.request.sendall(b'success\n')
        elif data == b'show':
            self.request.sendall(json.dumps(CLIENT_STATE.current_cart.show()).encode('utf-8') + b'\n')


class ThreadedTCPServer(socketserver.ThreadingMixIn, socketserver.TCPServer):
    pass


if __name__ == "__main__":
    HOST, PORT = "0.0.0.0", 9999

    # Create the server, binding to localhost on port 9999
    with ThreadedTCPServer((HOST, PORT), MyTCPHandler) as server:
        # Activate the server; this will keep running until you
        # interrupt the program with Ctrl-C
        server_thread = threading.Thread(target=server.serve_forever)
        server_thread.daemon = True
        server_thread.start()

        # periodically sync with server
        while True:
            CLIENT_STATE.sync_with_server()
            time.sleep(60)
