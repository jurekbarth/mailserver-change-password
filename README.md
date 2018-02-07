# Addon for tomav/docker-mailserver

## build image

docker build -t jurekbarth/gocrypt:latest .

## run image

docker run --name gocrypt -v PATH_TO_CONFIG_FOLDER:/app/config -p 8000:8000 jurekbarth/gocrypt
