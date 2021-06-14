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

    def to_json(self):
        return {"product_id": self.product_id,
                "product_name": self.product_name,
                "price": self.price,
                "quantity": self.quantity}

    @classmethod
    def from_json(cls, json_dict):
        return cls(**json_dict)


class ShoppingCart(object):
    def __init__(self):
        self._set = set()
        self._tombstone_set = set()

    def _make_id(self):
        return str(uuid.uuid4())

    def add(self, elem):
        self._set.add((self._make_id(), elem))

    def remove(self, elem):
        for elem_tuple in self._set:
            if elem_tuple[1] == elem:
                self._tombstone_set.add(elem_tuple)

    def show(self):
        current_set = self._set.copy()
        current_tombstone_set = self._tombstone_set.copy()
        effective_set = set([elem_tuple[1] for elem_tuple in current_set if elem_tuple not in current_tombstone_set])
        return [elem.to_json() for elem in effective_set]

    def get_set(self):
        return self._set

    def get_tombstone_set(self):
        return self._tombstone_set

    def to_json(self):
        # dumps shopping cart to a json serializable form
        current_set = self._set.copy()
        current_tombstone_set = self._tombstone_set.copy()
        return {"set": [[tuple_elem[0], tuple_elem[1].to_json()] for tuple_elem in current_set],
                "tombstone_set": [[tuple_elem[0], tuple_elem[1].to_json()] for tuple_elem in current_tombstone_set]}

    @classmethod
    def from_json(cls, json_dict):
        loaded_set = set([(list_elem[0], CartItem.from_json(list_elem[1])) for list_elem in json_dict["set"]])
        loaded_tombstone_set = set([(list_elem[0], CartItem.from_json(list_elem[1])) for list_elem in json_dict["tombstone_set"]])
        result = cls()
        result._set = loaded_set
        result._tombstone_set = loaded_tombstone_set
        return result
