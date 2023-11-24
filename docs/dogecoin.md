## Run dogecoin node

### 1. Download dogecoin program
```
https://github.com/dogecoin/dogecoin/releases
```


### 2. Start dogecoin node
```
Configuration file
server=1
rpcuser=admin
rpcpassword=admin
rpcallowip=0.0.0.0
txindex=1

Or start the command directly
./dogecoind -datadir="data" -txindex -server -rpcuser=admin -rpcpassword=admin -rpcallowip="0.0.0.0/0" -rpcbind=0.0.0.0:22555
```