#!/bin/bash

db(){
    POLICY=$1
    if [ $POLICY == "start" ]; then
        echo "gcloud sql instances patch simplusers --activation-policy ALWAYS";
        eval "gcloud sql instances patch simplusers --activation-policy ALWAYS";
    elif [ $POLICY == "finish" ]; then
        echo "gcloud sql instances patch simplusers --activation-policy NEVER";
        eval "gcloud sql instances patch simplusers --activation-policy NEVER";
    elif [ $POLICY == "restart" ]; then
        echo "gcloud sql instances restart simplusers";
        eval "gcloud sql instances restart simplusers";
    fi
}

"$@"