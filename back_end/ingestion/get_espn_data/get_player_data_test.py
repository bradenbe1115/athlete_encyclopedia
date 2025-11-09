import pytest
import get_player_data
from common.models import Athlete

@pytest.mark.parametrize(
    ("athlete_input", "expected"),
    [
        (Athlete(id="invalid_due_to_name", firstName="-", jersey="10"),
         False),

        (Athlete(id="valid", firstName="Player", jersey="10"),
         True
         ),

        (Athlete(id="invalid_due_to_jersey", firstName="Player", jersey="-"),
            False),

        (Athlete(id="invalid_due_to_jersey_number", firstName="Player", jersey="55"),
         False),

        (Athlete(id="valid_jersey_above_80", firstName="Player", jersey="81"),
          True)
    ]
)
def test_is_valid_player_data(athlete_input, expected):
    assert get_player_data.is_valid_player_data(athlete_input) == expected
