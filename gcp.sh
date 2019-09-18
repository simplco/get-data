#!/bin/bash

# get the gcloud sdk hur (cmd|ctrl + click): https://cloud.google.com/sdk/docs/downloads-versioned-archives

# start, stop restart cloud sql instance with ". gcp.sh db (start | stop | restart)"

db(){
    POLICY=$1
    if [ $POLICY == "start" ]; then
        echo "gcloud sql instances patch simplusers --activation-policy ALWAYS";
        eval "gcloud sql instances patch simplusers --activation-policy ALWAYS";
    elif [ $POLICY == "stop" ]; then
        echo "gcloud sql instances patch simplusers --activation-policy NEVER";
        eval "gcloud sql instances patch simplusers --activation-policy NEVER";
    elif [ $POLICY == "restart" ]; then
        echo "gcloud sql instances restart simplusers";
        eval "gcloud sql instances restart simplusers";
    fi
}

# create a cloud function with ". gcp.sh fn FUNCTION-NAME"
# e.g. ". gcp.sh fn Read"
# run "gcloud functions delete FUNCTION-NAME" to delete function

fn() {
    FNAME=$1
    REPO="https://source.developers.google.com/projects/testsite-234503/repos/simpl/moveable-aliases/CRUD/paths/cloud"
    CMD="gcloud functions deploy $FNAME --region=us-central1 --memory=256MB \
    --runtime=go111 --source=$REPO --env-vars-file=cloud/.env.yaml --trigger-http"

    echo $CMD
    eval $CMD
}
"$@"