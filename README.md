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
docker run -d -e AUTH_USERNAME=<your username> -e AUTH_PASSWORD=<your password> -v /data/post-handler:/root/data -p 80:80 dockerbluestone/post-handler:latest
```

### Docker-Compose
```yaml
version: "3.9"  # optional since v1.27.0
services:
  post-handler:
    image: dockerbluestone/post-handler:latest
    ports:
      - "80:80"
    volumes:
      - /data/post-handler:/root/data
    environment:
      - AUTH_USERNAME=username
      - AUTH_PASSWORD=password
    networks: 
      - post-handler

networks:
  post-handler:
    driver: bridge
```