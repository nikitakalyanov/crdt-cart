Set up MariaDB:

```
docker network create mysql-network

docker run \
  -d \
  --network mysql-network \
  -p 3306:3306 \
  -e MARIADB_ROOT_PASSWORD=testrootpassword \
  --name cart-mariadb mariadb:10.7.3
```

```
# fill cart initially
docker exec -it cart-mariadb mysql -ptestrootpassword
CREATE DATABASE crdtcart;
USE crdtcart;
CREATE TABLE cart (id INTEGER, current_cart LONGTEXT);
INSERT INTO cart VALUES (1, '[]');
```

Run cart filler:
```
go build .
# run a container at the same network with mounted dir
docker run --rm --name mariadb-runner --network=mysql-network -v /home/nkalyanov/crdt-cart/:/crdt-cart/ -it golang bash
/crdt-cart/go-crdt/mysql/mysql
```

```
# docker stop cart-mariadb
# docker rm cart-mariadb
```
