# APK Service Monitor
Service status monitor for debian linux distributions written in Go

* **This program needs sudo permission to run**

## Overview

APKSM performs checks if service is available every *t* seconds and notifies when unsuccessful


## Instruction

### Build and run

Run from terminal:

```bash
cd cmd
go build -o apksm
chmod +x apksm
./apksm -config=config.json
```

This command will build and run program with configuration file `configs/config.json`.

### Command line arguments

Following arguments are available:

`config` The configuration file path (default "configs/default.json")

`log` Log file store path (default "logs/{current-date}.log")

`host` The host address for http server (default "127.0.0.1")

`port` The host port for http server (default "8080")

`logfilter` The text to filter log by

`logging` If put this argument false, file logging will disable

Example: `./apksm -config=configs/myconfig.json -logfilter=nginx -host=192.168.0.1 -port=1010 -logging=true`.

Web interface will be available at `192.168.0.1:1010`.

File logs will be store at `/var/log/apksm`

You can also use `./apksm -help` for help.

## Configuration

JSON structure is used for configuration. Example can be found in `configs/default.json`.

```json
{
  "settings": {
    "notifications": {
      "email": [
        {
          "smtp": "smtp.gmail.com",
          "port": 587,
          "username": "example@gmail.com",
          "password": "******",
          "from": "example@gmail.com",
          "to": [
            "example@gmail.com"
          ]
        }
      ],
      "telegram": [
        {
          "botToken": "123456:ABC-DEF-...",
          "chatId": "99999999"
        }
      ],
      "webhook": [
        {
          "url": "url",
          "method": "GET"
        }
      ]
    },
    "monitor": {
      "checkInterval": 5,
      "CPUMax": 1,
      "monitorSystemUsage": true,
      "memoryMax": 1000,
      "exponentialBackoffSeconds": 5
    }
  },
  "services": [
    {
      "name":"nginx",
      "specificPattern":"high",
      "checkInterval": 5,
      "restartIfDown": true,
      "saveLogsOnStop": true
    }
  ]
}

```

### Global

`checkInterval` Check interval for each service or system usage

`monitorSystemUsage` If true, enables system usage monitoring

`CPUMax` The percentage of maximum CPU usage, if go upper then system will notify

`memoryMax` The amount of maximum memory usage in kiloBytes, if go upper then system will notify

`exponentialBackoffSeconds` After each notification, time until next notification is available is increased exponetially. On first unsuccessful service status reach, notification will always be sent immediately. If, for example, `exponentialBackoffSeconds` is set to `5`, then next notifications will be available after 5, 25, 125... seconds. On successful service status reach after downtime, this will be reset.  

### Services

`name` Systemd service name like `nginx`

`specificPattern` The pattern that system will try to find it in service log and if present in the logs it will notify

`restartIfDown` If true, system will restart this system if it finds it down

`saveLogsOnStop` If true, system will save the service logs after it find the service down

`checkInterval`  Check interval for each service in seconds (this will override global settings)


### Notifications

There can be multiple notification settings.

#### Email

`smtp` SMTP server address

`port` SMTP server port

`username` Login email

`password` Login password

`from` Email that notifications will be sent from

`to` Array of recipients 

#### Telegram

`botToken` Telegram Bot token obtained from the BotFather.

`chatId` Chat-ID of the user to message (It Can also be a group id).


#### Webhook

`url` url to make request to

`method` method to use (`GET` or `POST`)

Service information will be stored in `service` parameter.


## API
This application has API to retrieve status of each service in JSON format

API of current status is available at `/api` endpoint.

Example is given below.

```json
{
  "nginx": [
    {
      "time": "2023-12-05T20:51:34.384739279+03:30",
      "running": true
    },
    {
      "time": "2023-12-05T20:51:39.401586335+03:30",
      "running": true
    }
  ]
}
```
