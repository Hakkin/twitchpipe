#!/bin/bash

DEPS=("websocat" "curl" "jq" "cut" "twitchpipe")

API_URL="https://api.twitch.tv/kraken/channels/"
WEBSOCKET_URL="wss://pubsub-edge.twitch.tv/v1"
CLIENT_ID="jzkbprff40iqj646a697cyrvl0zt2m6"

errf(){ >&2 printf "$@"; }

get_id () {
  curl --silent --fail -H "Client-ID: $CLIENT_ID" "$API_URL$1" | jq -r ._id
}

check_deps() {
  for i in "${DEPS[@]}"
  do
    if ! [ -x "$(command -v ${i})" ]
    then
      errf "record: missing dependency '%s'\n" "${i}"
      exit 1
    fi
  done
}

check_deps

if [ "$#" -lt 1 ]; then
    errf 'Please supply username(s)'
    exit 1
fi

errf 'fetching user IDs...\n'

declare -A id_username
for username in "$@"
do
  id=$(get_id $username)
  if [ "$id" == "" ];
  then
    errf 'ERROR! could not get ID for "%s"' "$username"
    exit 1
  fi
  id_username[$id]=$username
done

(
  trap exit SIGINT SIGTERM
  while :
  do
    errf "connecting to websocket...\n"
    (
      (
        while :
        do
          echo '{"type":"PING"}'
          sleep 150
        done
      ) &
      for k in "${!id_username[@]}"
      do
        printf '{"type":"LISTEN","data":{"topics":["video-playback-by-id.%d"]}}\n' "$k"
      done
    ) | websocat "$WEBSOCKET_URL"
    errf "disconnected, re"
  done
) | (
  errf "monitoring streams...\n"
  while read -r line;
  do
    #errf ">%s\n" "$line"
    type=$(echo "$line" | jq -r .type)
    case $type in
    MESSAGE)
      id=$(echo "$line" | jq -r .data.topic | cut -d . -f 2)
      message=$(echo "$line" | jq -r .data.message)
      message_type=$(echo "$message" | jq -r .type)
      
      if [ "$message_type" == "stream-up" ];
      then
        (
          username=${id_username[$id]}
          errf "[%s] stream started\a\n" "$username"
          while :
          do
            mkdir -p "$username"
            filename=$(printf "%s/%s.ts" $username "$(date -u '+%Y_%m_%d_%H_%M_%S_(%Z)')")
            errf "[%s] recording to %s\n" "$username" "$filename"
            (twitchpipe --archive "$username" >> "$filename") 2>&1 | (
              while read -r line;
              do
                errf "[%s] %s\n" "$username" "$line"
              done
            )
            if [ "${PIPESTATUS[0]}" == "2" ];
            then
              errf "[%s] stream ended with error, restarting...\n" "$username"
              continue
            fi
            break
          done
        ) &
      fi
      ;;
    PONG)
      ;;
    RESPONSE)
      ;;
    *)
      errf ">%s\n" "$line"
      ;;
    esac
  done
)