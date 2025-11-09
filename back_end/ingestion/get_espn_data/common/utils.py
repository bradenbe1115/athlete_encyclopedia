import os
from pathlib import Path

def get_espn_league() -> str:
    try:
        return os.environ["ESPN_LEAGUE"]
    
    except Exception as e:
        raise KeyError("Missing required environment variable ESPN_LEAGUE", e)

def get_ath_encylopedia_home() -> str:
    try:
        return os.environ["ENC_HOME"]

    except Exception as e:
        raise KeyError("Missing required environment variable ENC_HOME", e)
    
def get_players_file() -> str:
    try:
        return os.environ["PLAYER_FILE"]
    
    except Exception as e:
        raise KeyError("Missing required environment variable PLAYER_FILE", e)
