FROM rust:latest AS builder

WORKDIR /opt/src/mail-organizer
COPY ./src/main.rs ./src/main.rs
COPY ./Cargo.toml ./Cargo.toml

RUN cargo build --release

FROM debian:buster

RUN mkdir -p /opt/etc/configs && apt-get update && apt-get -y upgrade && apt-get -y install openssl ca-certificates
COPY --from=builder /opt/src/mail-organizer/target/release/mail-organizer /opt/bin/mail-organizer

CMD /opt/bin/mail-organizer /opt/etc/configs/*.yaml
