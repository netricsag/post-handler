# post-handler
A webserver, which handles HTTP Post requests to safe it into a OS Path

## Env Variables
Set the following required Env Variables to start the HTTP Post Handler
```bash
export AUTH_USERNAME=<your username>
export AUTH_PASSWORD=<your password>
```
**optional** you can set another server port with the following Env Variable (default is Port 80)
```bash
export SERVER_PORT=<your port>
```

## Volumes
The program creates a folder called **"./data"** in the executing path and stores every file which its getting in this directory.

## Docker
```bash
docker run -e AUTH_USERNAME=<your username> -e AUTH_PASSWORD=<your password> -v /data:/app/data -p 80:80
```