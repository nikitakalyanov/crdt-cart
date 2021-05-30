from crdt_cart import common

class ClientCartSyncronizer(common.BaseCartSyncronizer):
    def sync_with_server(self):
        print('no-op sync with server')
