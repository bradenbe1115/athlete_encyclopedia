from pydantic import BaseModel
from typing import Optional

class ListAthletesResponse(BaseModel):
    count: int
    pageIndex: int
    pageSize: int
    pageCount: int
    items: list

class AthleteInfoResponse(BaseModel):
    statisticslog: dict = None
    position: dict = None

class AthleteStatsLogResponse(BaseModel):
    entries: list

class Athlete(BaseModel):
    id: str
    firstName: str = "-"
    jersey: str = "-"