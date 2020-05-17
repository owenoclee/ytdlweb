# ytdlweb

A little web front-end for youtube-dl. Note this tool only downloads the video
to the server, not the client. In my case I like to use Nextcloud to synchronise
the videos to my local machines.

## Running

Basic example:

```bash
docker build -t ytdlweb .
docker run --name ytdlweb -p 80:3000 -v ~/ytdl:/ytdlweb/downloads -d ytdlweb
```

This will start the web server on `http://localhost` and provide the host
machine access to the downloads at `~/ytdl`.

### Optional Configuration

If you like, you can configure the following environment variables:

- `LISTEN_ADDR` - the address that the http server will listen on.
  Default: `:3000`.
- `WORKER_COUNT` - the max no. of videos that can be downloaded simultaneously.
  Default: `4`.

## Disclaimer

This is a personal project that I provide no guarantees for. It is certainly not
"production ready". At minimum I'd recommend running it behind a reverse proxy
with basic authentication and only granting access to trustworthy friends.
