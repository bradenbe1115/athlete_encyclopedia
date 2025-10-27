import logging
import scrapy_runner

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

if __name__ == "__main__":
    scrapy_runner.run(logger)
