# get_espn_data

This module contains a client for interacting with the hidden ESPN API to get player data.

You can run the module locally via uv or in a Docker container. 

To run in docker, make sure to create the lock file first via `make lock-deps`.

You can build the docker image with:
`docker build -f ./back_end/ingestion/get_espn_data/Dockerfile -t get-espn-data .`

You can then execute a runner with:
`docker run -e RUNNER=get_player_data -e ESPN_LEAGUE=college-football -v $ENC_HOME:/athlete_encyclopedia  get-espn-data get_players.py`