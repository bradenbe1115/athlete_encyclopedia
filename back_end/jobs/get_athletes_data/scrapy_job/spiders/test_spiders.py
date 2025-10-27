import pytest
import scrapy
from spiders.sr_spiders import process_athlete_list_page


@pytest.mark.unit
@pytest.mark.parametrize(
    ("input_html_file", "expected"),
    [
        # Valid/Expected HTML 
        (
            "scrapy_job/spiders/test_data/sports_reference_athlete_div/normal.html",
            [
                {
                    "name": "Terry Anthony",
                    "team": "Florida State",
                    "years": " (1986-1989) ",
                },
                {"name": "Tom Anthony", "team": "Ohio State", "years": " (1982-1982) "},
            ],
        ),
        # HTML with one valid athlete and one invalid
        (
            "scrapy_job/spiders/test_data/sports_reference_athlete_div/missing_team.html",
            [
                {
                    "name": "Terry Anthony",
                    "team": "Florida State",
                    "years": " (1986-1989) ",
                },
            ],
        ),
        # Invalid HTML
        (
            "scrapy_job/spiders/test_data/sports_reference_athlete_div/invalid.html",
            [],
        ),
    ],
)
def test_process_athlete_list_page(input_html_file, expected):
    with open(input_html_file, "r", encoding="utf-8") as f:
        html_content = f.read()

    resp = scrapy.http.TextResponse(url="http://x", body=html_content, encoding="utf-8")

    results = process_athlete_list_page(resp)
    assert results == expected
