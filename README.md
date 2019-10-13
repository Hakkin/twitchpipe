# twitchpipe
Pipe your favorite Twitch streams to the media player of your choice, or a file to save them for later.

Supports low-latency playback.

# Installation
```
go get github.com/Hakkin/twitchpipe
```

# Usage
```
Usage: twitchpipe [OPTIONS...] <USERNAME> [COMMAND...]

If COMMAND is specified, it will be executed and stream data will be
written to its standard input.
Otherwise, stream data will be written to standard output.

Options:
  -a, --archive
        Start downloading from the oldest segment rather than the newest
  -f, --force-output
        Force output to standard output even if TTY is detected
  -h, --hide-console
        Hide own console window
  -u, --url
        Treat USERNAME as a URL
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