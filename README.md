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

**optional** you can set some SMB users to push the files directly on a SMB Share
```bash
export SMB_ENABLED=true
export SMB_SERVER=<IP or DNS> # 192.168.1.10
export SMB_SHARENAME=<sharename> # The name of the Windows Share (not \\192.168.1.10\share, only share)
export SMB_USERNAME=<smb username> # without domain
export SMB_PASSWORD=<smb password>
export SMB_DOMAIN=<windows domain> # e.g. domain.local
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
      - SMB_ENABLED=true
      - SMB_SERVER=192.168.1.10
      - SMB_SHARENAME=share
      - SMB_USERNAME=username
      - SMB_PASSWORD=password
      - SMB_DOMAIN=domain.local
    networks: 
      - post-handler

networks:
  post-handler:
    driver: bridge
```

