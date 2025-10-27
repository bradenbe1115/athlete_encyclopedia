import scrapy


def process_athlete_list_page(response) -> list[dict]:
    """
    Processes athlete list page from Sports Reference and returns data for each athlete.

    If a either the team or year field is missing, the athlete will be skipped when processing.

    Args:
        response: HTML response from Sports Reference

    Returns:
        result (list[dict]): list of dicts containing scraped data for all athletes on page
    """

    results = []
    for athlete in response.xpath("//div[@id='div_players']/p"):
        result = {}
        athlete_anchor_tags_text_list = athlete.css("a::text").getall()
        div_text_list = athlete.css("::text").getall()

        # skip if data is not in expected format
        if len(athlete_anchor_tags_text_list) < 2 or len(div_text_list) < 4:
            break

        result["name"] = athlete_anchor_tags_text_list[0]
        result["team"] = athlete_anchor_tags_text_list[1]
        result["years"] = div_text_list[3]

        results.append(result)

    return results


class CFPSpider(scrapy.Spider):
    name = "cfb_athletes"
    root_url = "https://www.sports-reference.com/cfb/players"

    async def start(self):
        all_letters = list(string.ascii_lowercase)

        # one start page for each letter
        urls = [f"{self.root_url}/{x}-index.html" for x in all_letters[0:2]]

        for url in urls:
            yield scrapy.Request(url=url, callback=self.parse)

    def parse(self, response):
        results = process_athlete_list_page(response)
        for result in results:
            yield result

        next_page = response.xpath("//a[@class='button2 next']/@href").get()
        if next_page is not None:
            next_page = response.urljoin(next_page)
            yield scrapy.Request(next_page, callback=self.parse)
