# mail-organizer

I wrote this small utility to clean up my mail inboxes. It is planned to let this utility run with a cron job, so it scans my mails every 5 minutes and moves those away based on my configured rules.

This utility is not stable yet. It's under heavy development and might change completely.

## Prerequisites

- Go toolchain installed
- Mail account with `UIDPLUS` capability

## Install

1. First clone this repository.
2. To just build this utility:
```bash
go build .
```

## Usage

Execute the program and add your configs as arguments.

```bash
./mail-organizer ./example_config.yaml ./example_config_2.yaml
# Or without prior building
go run . ./example_config.yaml ./example_config_2.yaml
```

## Using Docker

First create a docker image:
```bash
docker build -t mail-organizer:0.1.0 .
```

Next run the Docker Image and mount your configs:
```bash
docker run --rm -t -v $(PWD)/example.yaml:/home/nonroot/example.yaml:ro mail-organizer:0.1.0
```
