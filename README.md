# wg-logs

GO program that checks every minute which wireguard connections are up and writes them to a sqlite database

## TODO

- Close entry if last handshake or transfer fields are gone
- Make linux service
- Implement option to forward events

## Build

To run this project:

```
go mod tidy
go run . [svc|sync|sync-test|show]
```

To compile:

```
bash scripts/build.sh
```
