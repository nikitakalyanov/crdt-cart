import json
from http.server import HTTPServer, BaseHTTPRequestHandler
import threading

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
            client_state = ShoppingCart.from_json(dict_client_state)
            SERVER_STATE.merge(client_state)

            # async save to db
            save_thread = threading.Thread(target=SERVER_STATE.save_to_db)
            save_thread.daemon = True
            save_thread.start()

            self._set_headers()
            self.wfile.write(json.dumps({"success": True,
                                         "server_state": SERVER_STATE.current_cart.to_json()}).encode('utf-8'))

    def do_GET(self):
        if self.path == '/state':
            print('getting current server state')
            self.send_response(200)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(SERVER_STATE.current_cart.show()).encode(
                'utf-8'))


def main(server_class=HTTPServer, handler_class=MyHandler, addr="localhost", port=80):
    server_address = (addr, port)
    http = server_class(server_address, handler_class)

    print(f"Starting http server on {addr}:{port}")
    http.serve_forever()


if __name__ == "__main__":

    main(addr='0.0.0.0', port=12347)
