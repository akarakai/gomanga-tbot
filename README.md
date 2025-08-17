# gomanga-tbot

Simple Telegram bot that downloads manga from [WeebCentral](https://weebcentral.com/) using Playwright and notifies users when a new chapter of their favorite manga is available.

## How to run
The easiest way to run this bot is by running it with the docker.

Build your own image from the `Dockerfile` or pull it from docker hub

```bash
docker pull akarakaii/gomanga:latest
```

Then run the container
```bash
docker run -d \
  --name gomanga-telegram-bot \
  -e TELEGRAM_API_KEY=your_api_key_here \
  -v path_to_host/database.db:/app/database.db \
  akarakaii/gomanga:latest
```

At this moment only Sqlite is supported as database