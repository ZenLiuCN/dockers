# pgpool2 docker image on alpine
## usage
pgpool2 version `3.7.9`
docker image `lcz20/zen:pgpool2`
## example
```yaml
version: '3'
services:
  pg1:
    image: groonga/pgroonga:2.2.0-alpine-slim-11
    ports:
      - 5430:5432
  pg2:
    image: groonga/pgroonga:2.2.0-alpine-slim-11
    ports:
      - 5431:5432
  pg3:
    image: groonga/pgroonga:2.2.0-alpine-slim-11
    ports:
      - 5433:5432
  pgpool2:
    image: lcz20/zen:pgpool2
    depends_on:
      - pg1
      - pg2
      - pg3
    environment:
      - PGPOOL_BACKENDS=1:pg1:5432,2:pg2:5432,3:pg3:5432
      - PGPOOL_REPLICATION_MODE=1
    ports:
      - 5432:5432/tcp
```
