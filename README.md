# wg-logs

GO program that checks every minute which wireguard connections are up and writes them to a sqlite database

## Install

### Copy binary

Place wg-logs in /usr/local/bin, make root the owner and fix permissions

```
sudo mv wg-logs /usr/local/bin
sudo chown root:root /usr/local/bin/wg-logs
sudo chmod 700 /usr/local/bin/wg-logs
```

### Create lib folder

```
sudo mkdir /usr/local/bin/wg-logs
sudo chmod 700 /usr/local/bin/wg-logs
```

### systemd service

Place the following config in `/etc/systemd/system/wg-logs.service`

```
[Unit]
Description=WireGuard Log Collector (wg-logs)
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/var/lib/wg-logs
ExecStart=/usr/local/bin/wg-logs svc
Restart=on-failure
RestartSec=5s
KillSignal=SIGINT
TimeoutStopSec=20
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Now enable en start the service

```
sudo systemctl daemon-reload
sudo systemctl enable wg-logs
sudo systemctl start wg-logs
```

## Show events

Run `wg-logs show`

## TODO

- Close entry if last handshake or transfer fields are gone
- Make linux service install script
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
