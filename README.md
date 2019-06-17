ubiquiti-mfi-influxdb-go
================

What does it do?
----------------

Reads power statistics from mFi devices and sends them to an InfluxDB instance at defined intervals. 

Usage
-----

Available flags:

```cgo
$ ./ubiquiti-mfi-influxdb-go
Usage of ./ubiquiti-mfi-influxdb-go:
  -mfi-addr string
    	address of the mFi device (default "http://127.0.0.1")
  -mfi-user string
    	username to access the mFi device (default "ubnt")
  -mfi-pass string
    	password to access the mFi device (default "ubnt")
  -http-port string
    	port for the http server to listen on for health checks (default "8085")
  -influxdb-addr string
    	address of influxdb endpoint, ex: http://127.0.0.1:8086 (default "http://127.0.0.1:8086")
  -influxdb-database string
    	influxdb database to store datapoints (default "homelab_custom")
  -influxdb-pass string
    	password for influxdb access (default "admin")
  -influxdb-user string
    	username for influxdb access (default "admin")

```

To Do
------

- Add real error handling
- Make the interval customizable