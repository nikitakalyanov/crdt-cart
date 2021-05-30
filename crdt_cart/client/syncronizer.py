import requests

from crdt_cart import common
from crdt_cart.common.cart import ShoppingCart


class ClientCartSyncronizer(common.BaseCartSyncronizer):
    def __init__(self, server_url):
        super(ClientCartSyncronizer, self).__init__()
        self.server_url = server_url

    def sync_with_server(self):
        print('syncing with server')
        server_response = requests.post(self.server_url, json=self.current_cart.to_json())
        server_state_dict = server_response.json()['server_state']
        server_state = ShoppingCart.from_json(server_state_dict)
        self.merge(server_state)
