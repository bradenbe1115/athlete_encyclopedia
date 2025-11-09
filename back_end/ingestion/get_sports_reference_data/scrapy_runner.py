import os
import logging
import subprocess
from datetime import datetime

def run(logger: logging.Logger):
    """
        Runs scrapy spider and saves output to jsonl file.

        Saves the file name as the spider name + current datetime.
    """
    spider_name = os.getenv("SPIDER_NAME")
    spider_name="cfb_athletes"
    if spider_name is None:
        raise Exception("Missing required Env Var SPIDER_NAME")

    scrape_dt = datetime.now().strftime("%d-%m-%Y-%H-%M-%S")

    output_file_name = f"{spider_name}_{scrape_dt}.jsonl"
    logger.info(f"Running scrapy spider {spider_name} at datetime {scrape_dt}.")
    subprocess.run(["scrapy", "crawl", spider_name, "-o", output_file_name])
