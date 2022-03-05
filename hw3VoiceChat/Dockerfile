FROM python:3
EXPOSE 8080

WORKDIR /voice

COPY server.py ./
COPY protocol.py ./

ENTRYPOINT python3 server.py