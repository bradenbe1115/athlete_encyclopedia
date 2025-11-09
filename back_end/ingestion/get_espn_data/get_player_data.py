import sys
import logging
from datetime import datetime

from common.espn_client import ESPNClient
from common.json_writer import write_to_JSON, read_from_JSON
from common.models import Athlete
from common.options import LEAGUE_OPTIONS
from common.utils import get_ath_encylopedia_home, get_espn_league, get_players_file

RESULT_FILE_NAME_TEMPLATE = "/{league}/players_{datetime}.json"

def is_valid_player_data(ath: Athlete) -> bool:
    """
        Determines if the athlete is valid for details ingestion.

        Invalide athlete records are missing a jersy number or have a malformed name only hyphens in the title.

        TEMP: filtering out athletes with a numbers between 50 & 80
    """

    if ath.firstName.lower() in ['-','team']:
        return False
    
    if ath.jersey == '-':
        return False
    
    if int(ath.jersey) > 49 and int(ath.jersey) < 80:
        return False
    
    return True

def run():
    logger = logging.getLogger()
    logger.setLevel(logging.INFO)
    handler = logging.StreamHandler(sys.stdout)
    handler.setLevel(logging.DEBUG)

    formatter = logging.Formatter(
        "%(asctime)s [%(levelname)s] %(name)s: %(message)s"
    )
    handler.setFormatter(formatter)

    logger.addHandler(handler)


    league = get_espn_league()
    ath_encylopedia_home = get_ath_encylopedia_home()
    players_file = get_players_file()
    
    sport = LEAGUE_OPTIONS[league]["sport"]
    logger.debug(f"Found sport: {sport} associated with league {league}.")

    client = ESPNClient(logger=logger)
    players = read_from_JSON(players_file)
    athlete_objs = [Athlete(**player) for player in players]

    filtered_athlete_objs = [ath_obj for ath_obj in athlete_objs if is_valid_player_data(ath_obj)]
    
    results = client.get_players_details(filtered_athlete_objs[0:2], sport ,league)

    current_datetime = datetime.now().strftime("%d-%m-%Y-%H-%M%S")
    result_file_name = RESULT_FILE_NAME_TEMPLATE.format(league=league.replace("-","_"), datetime=current_datetime)
    write_to_JSON(results, f"{ath_encylopedia_home}/data/player_details/{result_file_name}")

if __name__ == "__main__":
    run()
