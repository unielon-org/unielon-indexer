﻿# unielon-indexer

### Because the file is relatively large, file pull is required.

```
1. sudo apt-get install git-lfs 
2. git lfs pull
```

## start up

### 1. Install and run dogecoin
Please check out [dogecoin](docs/dogecoin.md)

### 2. Compile golang program
```go
go build.
```

### 3. Download data

https://github.com/unielon-org/unielon-indexer/releases

Download the latest db data of releases and put it in the data directory



### 4. Run
```go
./unielon-indexer
```

### Router Document
Please check out [router](https://documenter.getpostman.com/view/8337528/2s9YeN18PF)
