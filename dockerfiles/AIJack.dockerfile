FROM pytorch/pytorch:latest

COPY ./AIJack /tmp/AIJack

COPY ./.condarc /root/.condarc

COPY ./start.sh /tmp/start.sh

RUN sh /tmp/start.sh
