# mail-organizer

I wrote this small utility to clean up my mail inboxes. It is planned to let this utility run with a cron job, so it scans my mails every 5 minutes and moves those away based on my configured rules.

This utility is not stable yet. It's under heavy development and might change completly.

## Prerequisites

- Rust toolchain installed
- Mail account with `UIDPLUS` capability (Attention: This is currently not checked before trying to apply the rules!!)

## Install

1. First clone this repository.

- To just build this utility:
```bash
cargo build --release
```
- To install
```bash
cargo install
```

## Usage

Execute the program and add your configs as arguments.

```bash
./target/release/mail-organizer ./example_config.yaml ./example_config_2.yaml
# Or without prior building
cargo run --release -- ./example_config.yaml ./example_config_2.yaml
```
