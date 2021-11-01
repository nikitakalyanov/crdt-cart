import json
from http.server import HTTPServer, BaseHTTPRequestHandler
import sys
import threading

import requests
from requests import codes

from crdt_cart.server import ServerCartSyncronizer
from crdt_cart.common.cart import ShoppingCart


SERVER_STATE = ServerCartSyncronizer()


class MyHandler(BaseHTTPRequestHandler):
    def _set_headers(self):
        self.send_response(200)
        self.send_header("Content-type", "application/json")
        self.end_headers()

    def do_POST(self):
        print('got POST request for %s' % self.path)
        if self.path == '/sync_state':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)

            dict_client_state = json.loads(post_data)

            # write to primary server
            server_response = requests.post(self.server.primary_crdt_endpoint,
                                            json=dict_client_state)
            if server_response.status_code == codes.OK:
                success = True
            else:
                success = False
                print('primary server returned unexpected response code',
                      server_response.status_code)

            self._set_headers()
            self.wfile.write(json.dumps({"success": success,
                                         "server_state": SERVER_STATE.current_cart.to_json()}).encode('utf-8'))

    def do_GET(self):
        if self.path == '/state':
            print('getting current server state')
            self.send_response(200)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(SERVER_STATE.current_cart.show()).encode(
                'utf-8'))


class PrimaryAwareServer(HTTPServer):
    def __init__(self, server_address, handler_class, primary_endpoint):
        super(PrimaryAwareServer, self).__init__(server_address, handler_class)
        self.primary_crdt_endpoint = primary_endpoint


def main(server_class=PrimaryAwareServer, handler_class=MyHandler,
         addr="localhost", port=80, primary_endpoint=None):

    server_address = (addr, port)
    http = server_class(server_address, handler_class, primary_endpoint)

    print(f"Starting http server on {addr}:{port}")
    http.serve_forever()


if __name__ == "__main__":

    args = sys.argv[1:]
    if len(args) != 2 or args[0] != '--primary':
        print('Incorrect usage, pass --primary primary-server:port/sync_state')
        sys.exit(1)

    main(addr='0.0.0.0', port=12347, primary_endpoint=args[1])
