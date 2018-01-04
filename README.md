# home-automation
Raspberry PI based home automation stuff

# snapshot
Snapshot is a `grpc` based service where multiple Raspberry PI Zero W clients communicate
with a central Raspberry PI 3 server. The central RPI 3
server has credentials to upload data received from clients
to the google storage bucket

## installation
```bash
# download code
go get -u github.com/sdeoras/home-automation/...
go generate github.com/sdeoras/home-automation/snapshot/client
go generate github.com/sdeoras/home-automation/snapshot/server
```
This will generate ARM-7 binary called `snapshot-server` and ARM-6 binary called
`snapshot-client` located at $GOPATH/bin

## starting the server
Copy `snapshot-server` to RPI 3 and make sure you have already downloaded credentials file
from GCP
```bash
export GOOGLE_APPLICATION_CREDENTIALS="path/to/gcp/credentials/file.json"
snapshot-server --bucket my-bucket &
```

## running client
Copy `snapshot-client` to RPI Zero W and make sure camera module is turned on.
```bash
snapshot-client --host server.ip.address --tag my-tag
```
If all goes well you will see a success message. You can check on your GCP account if the data
was uploaded successfully.

## setting up cron job
First define a bash script that calls client
```bash
#!/bin/bash
/path/to/snapshot-client --host server.ip.address --tag mytag >> /tmp/monitor.log
```
Then add crontab entry using `crontab -e`. Here client will execute every 30 minutes
```bash
*/30 * * * * /path/to/monitor.sh
```