from crdt_cart.common.cart import ShoppingCart

class BaseCartSyncronizer(object):
    def __init__(self):
        self.current_cart = ShoppingCart()

    def merge(self, target_cart: ShoppingCart):
        self.current_cart._set = self.current_cart._set + target_cart.get_set()
        self.current_cart._tombstone_set = self.current_cart._tombstone_set + target_cart.get_tombstone_set()
