import sys
import logging
from datetime import datetime

from common.espn_client import ESPNClient
from common.json_writer import write_to_JSON
from common.options import LEAGUE_OPTIONS
from common.utils import get_ath_encylopedia_home, get_espn_league

RESULT_FILE_NAME_TEMPLATE = "{league}/players_{datetime}.json"

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
    
    sport = LEAGUE_OPTIONS[league]["sport"]
    logger.debug(f"Found sport: {sport} associated with league {league}.")

    client = ESPNClient(logger=logger)
    players = client.get_players(sport, league)

    current_datetime = datetime.now().strftime("%d-%m-%Y-%H-%M%S")
    result_file_name = RESULT_FILE_NAME_TEMPLATE.format(league=league.replace("-","_"), datetime=current_datetime)
    write_to_JSON(players, f"{ath_encylopedia_home}/data/players/{result_file_name}")

if __name__ == "__main__":
    run()
