#!/bin/bash

DEPS=("websocat" "curl" "jq" "cut" "twitchpipe")

API_URL="https://gql.twitch.tv/gql"
WEBSOCKET_URL="wss://pubsub-edge.twitch.tv/v1"
CLIENT_ID="kimne78kx3ncx6brgo4mv6wki5h1ko"

PRINT_FILENAME=0
GROUP="chunked"

errf(){ >&2 printf "$@"; }

get_ids () {
  USERNAMES="$(printf "%s\n" "${@}")"
  USERNAMES_ARRAY="$(jq -R . <<< "${USERNAMES}" | jq -s -c .)"
  QUERY_STRING="$(printf "{users(logins:%s){id}}" "${USERNAMES_ARRAY}")"
  DATA="$(jq -c -R '{"query":.}' <<< "${QUERY_STRING}")"
  IDS="$(curl --silent --fail -H "Client-ID: ${CLIENT_ID}" "${API_URL}" --data-raw "${DATA}" | jq -r ".data.users[].id | .//-1")"
  echo -n "${IDS}"
}

print_usage() {
  errf "Usage: record [OPTIONS...] <USERNAMES...>\n"
  errf "\n"
  errf "Options:\n"
  errf "  -h\tPrints this help text\n"
  errf "  -p\tPrint filenames to standard output once stream ends\n"
  errf "  -g\tSelect playlist group to record\n"
}

check_deps() {
  for i in "${DEPS[@]}"
  do
    if ! [ -x "$(command -v ${i})" ]
    then
      errf $'missing dependency \'%s\'\n' "${i}"
      exit 1
    fi
  done
}

invalid_input() {
  errf "%s\n" "$1"
  errf $'try \'record -h\' for usage information'
  exit 1
}

check_deps

while getopts ":phg:" opt
do
  case "$opt" in
    p )
      PRINT_FILENAME=1
      ;;
    h )
      print_usage
      exit 0
      ;;
    g )
      GROUP="$OPTARG"
      ;;
    \? )
      invalid_input "$(printf $'unknown option \'-%s\'' "$OPTARG")"
      ;;
  esac
done

shift $((OPTIND -1))

USERNAMES=("${@,,}")

if [ "${#USERNAMES[@]}" -lt 1 ]; then
  invalid_input 'no username(s) supplied'
fi

errf 'fetching user IDs...\n'
IDS="$(get_ids "${USERNAMES[@]}" | tr -d '\r')"

declare -A id_username
i=0
while read -r ID
do
  USERNAME="${USERNAMES[${i}]}"
  if [[ "${ID}" -eq "-1" ]]
  then
    errf 'ERROR! could not get ID for "%s"' "$USERNAME"
    exit 1
  fi
  id_username["${ID}"]="${USERNAME}"
  ((i++))
done <<< "${IDS}"

(
  trap exit SIGINT SIGTERM
  while :
  do
    errf 'connecting to websocket...\n'
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
    errf 'disconnected, re'
  done
) | (
  errf 'monitoring streams...\n'
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
          errf '[%s] stream started\a\n' "$username"
          while :
          do
            mkdir -p "$username"
            filename=$(printf "%s/%s.ts" $username "$(date -u '+%Y_%m_%d_%H_%M_%S_(%Z)')")
            errf '[%s] recording to %s\n' "$username" "$filename"
            (twitchpipe --archive --group "$GROUP" "$username" >> "$filename") 2>&1 | (
              while read -r line;
              do
                errf '[%s] %s\n' "$username" "$line"
              done
            )
            EXIT_STATUS="${PIPESTATUS[0]}"
            if [ -s "${filename}" ]
            then
              if [ "$PRINT_FILENAME" == "1" ];
              then
                printf '%s\n' "$filename"
              fi
            else
              errf '[%s] output file %s was empty, removing...\n' "${username}" "${filename}"
              rm "${filename}"
            fi
            if [ "$EXIT_STATUS" == "2" ];
            then
              errf '[%s] stream ended with error, restarting...\n' "$username"
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
      errf '>%s\n' "$line"
      ;;
    esac
  done
)