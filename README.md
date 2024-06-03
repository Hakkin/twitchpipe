# twitchpipe
Pipe your favorite Twitch streams to the media player of your choice, or a file to save them for later.

Supports low-latency playback.

# Installation
```
go install github.com/Hakkin/twitchpipe@latest
```
Requires at least Go 1.18

# Usage
```
Usage: twitchpipe [OPTIONS...] <USERNAME> [COMMAND...]

If COMMAND is specified, it will be executed and stream data will be
written to its standard input.
Otherwise, stream data will be written to standard output.

Options:
  -G, --list-groups
        List available playlist groups and exit
  -a, --archive
        Start downloading from the oldest segment rather than the newest
  --access-token-device-id value
        Device ID to send when acquiring an access token (optional)
        If no device ID is specified, one will be generated randomly
  --access-token-oauth value
        OAuth token to send when acquiring an access token (optional)
  --access-token-platform string
        The platform to send when acquiring an access token (default "web")
  --access-token-player-backend value
        The player backend to send when acquiring an access token (optional)
  --access-token-player-type string
        The player type to send when acquiring an access token (default "site")
  -f, --force-output
        Force output to standard output even if TTY is detected
  -g, --group string
        Select specified playlist group
        "best" will select the best available group (default "best")
  -h, --hide-console
        Hide own console window
  -u, --url
        Treat USERNAME as a URL
  -v, --version
        Show version information and exit
```
`-h, --hide-console` is  a Windows specific switch that will hide the command prompt if `twitchpipe` is started directly.

### Example Usage
* Open stream `username` using `mpv`
  ```
  $ twitchpipe username mpv -
  ```
  
  alternatively, you can use a pipe
  
  ```
  $ twitchpipe username | mpv -
  ```
  
  the same can be done with most media players
  
  ```
  $ twitchpipe username vlc -
  ```
* Record stream `username` to `recording.ts`
  ```
  $ twitchpipe -a username > recording.ts
  ```
  `-a, --archive` will record starting from the oldest visible segment, useful for recording streams.
* Usernames can also be passed as a URL
  ```
  $ twitchpipe -u https://twitch.tv/username mpv -
  ```
  This can be useful for opening a stream from a web browser.