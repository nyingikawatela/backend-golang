FROM debian:stable-slim

COPY Chirpy-clone Chirpy-clone

ARG DB_URL
ARG PLATFORM

ENV DB_URL=""
ENV PLATFORM=""

EXPOSE 8080

CMD ["./Chirpy-clone"]