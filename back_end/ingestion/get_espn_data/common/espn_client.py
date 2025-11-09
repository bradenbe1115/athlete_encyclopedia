import requests
import logging
import common.models as models

class ESPNClient:
    BASE_URL = "https://sports.core.api.espn.com"

    def __init__(self, logger: logging.Logger):
        self.headers = {"Content-Type":"application/json"}
        self.logger = logger

    def _get(self, endpoint_url: str, params: dict) -> dict:
        resp = requests.get(endpoint_url, params=params, headers=self.headers)

        try:
            resp.raise_for_status()
            return resp.json()
        
        except Exception as e: 
            self.logger.info(f"Request raised exception {e}")

    def get_players(self, sport: str, league: str, results: list = None, limit: int = 18000, page: int = 1) -> list[dict]:
        endpoint_url = f"{self.BASE_URL}/v3/sports/{sport}/{league}/athletes"
        params = {"limit": limit, "page": page}

        self.logger.info(f"Get Players at endpoint {endpoint_url} with params {params}")
        resp = self._get(endpoint_url=endpoint_url, params=params)
        result = models.ListAthletesResponse(**resp)

        self.logger.debug(f"Retrieved {len(result.items)} athletes.")
        if results is None:
            results = result.items
        else:
            results += result.items

        self.logger.info(f"Retrieved {len(results)} total items.")
        
        next_page = page + 1
        self.logger.debug(f"Retrieved page {page} out of {result.pageCount} total pages.")
        if result.pageCount == next_page:
            return results

        return self.get_players(sport=sport, league=league, results=results, limit=limit, page=next_page) 
    
    def get_player_info(self, sport: str, league: str, id: str) -> dict:
        endpoint_url = f"{self.BASE_URL}/v2/sports/{sport}/leagues/{league}/athletes/{id}"
        
        self.logger.debug(f"Get player info at endpoint: {endpoint_url}")
        resp = self._get(endpoint_url, {})

        if resp is None:
            self.logger.debug(f"Endpoint {endpoint_url} returned no data")
        
        return resp
    
    def get_player_stats_log(self, sport: str, league: str, id: str) -> list:
        endpoint_url = f"{self.BASE_URL}/v2/sports/{sport}/leagues/{league}/athletes/{id}/statisticslog"

        self.logger.debug(f"Get player stats at endpoing: {endpoint_url}")
        resp = self._get(endpoint_url, {})

        result = models.AthleteStatsLogResponse(**resp)
        
        return result.entries

    def get_players_details(self, ath_objs: list[models.Athlete],sport: str, league: str, exceptions_collected: list[Exception] = [], exception_limit: int = 3) -> dict:
        """
            Orchestrates calls to player info and stats log for a list of player ids
             and  returns complete data from both endpoints for each player.

             Keeps track of number of exceptions caught due to unexpected behavior from API and returns results
             up to that point when limit is reached.
        """
        results = []
        for ath in ath_objs:
            try:
                id = ath.id
                
                self.logger.info(f"Getting player details for id: {id}")
                player_info = self.get_player_info(sport, league, id)
                ath_info_resp = models.AthleteInfoResponse(**player_info)

                if ath_info_resp.statisticslog is not None:
                    player_stats = self.get_player_stats_log(sport, league, id)
                    player_info["statsLog"] = player_stats
                
                results += [player_info]
            
            except Exception as e:
                exceptions_collected.append(e)
                
                if len(exceptions_collected) == exception_limit:
                    self.logger.info("Max exceptions reached in get players details. {exceptions_collected}")
                    return results

        return results            
    