import json

import psycopg2

from crdt_cart import common


class ServerCartSyncronizer(common.BaseCartSyncronizer):
    def __init__(self):
        super(ServerCartSyncronizer, self).__init__()
        self._db_connection = psycopg2.connect(
            database="crdtcart",
            user="server",
            password="server",
            host="127.0.0.1",
            port=5432,
        )
        self._db_connection.autocommit = True

    def save_to_db(self):
        print('saving to db')
        cursor = self._db_connection.cursor()
        cursor.execute("UPDATE cart SET current_cart = (%(current_cart)s) WHERE id = 1",
                       {"current_cart": json.dumps(self.current_cart.to_json())})
