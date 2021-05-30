import typing
import uuid


class CartItem(object):
    def __init__(self, product_id: int, product_name: str, price: float, quantity: int):
        self.product_id = product_id
        self.product_name = product_name
        self.price = price
        self.quantity = quantity

    def __eq__(self, other):
        return self.product_id == other.product_id

    def __hash__(self):
        return hash(self.product_id)

    def __str__(self):
        return "%s(%s)X%s" % (self.product_name, self.product_id, self.quantity)

    def __repr__(self):
        return "CartItem(product_id=%s, product_name='%s', price=%s, quantity=%s)" % (self.product_id, self.product_name, self.price, self.quantity)


class ShoppingCart(object):
    def __init__(self):
        self._set = set()
        self._tombstone_set = set()

    def _make_id(self):
        return uuid.uuid4()

    def add(self, elem):
        self._set.add((self._make_id(), elem))

    def remove(self, elem):
        for elem_tuple in self._set:
            if elem_tuple[1] == elem:
                self._tombstone_set.add(elem_tuple)

    def show(self):
        return set([elem_tuple[1] for elem_tuple in self._set if elem_tuple not in self._tombstone_set])

    def get_set(self):
        return self._set

    def get_tombstone_set(self):
        return self._tombstone_set
