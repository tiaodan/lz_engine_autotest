#!/bin/bash

FILE="aa.log"
INTERVAL=60
CMD="./replayer -f $FILE"

while true; do
    echo "$(date '+%F %T') 杀掉已有的 replayer 进程..."
    pkill -f "replayer -f"   # 杀掉匹配的进程

    echo "$(date '+%F %T') 启动新的 replayer..."
    $CMD &

    sleep $INTERVAL
done

