import sys
import asyncio
import json
import requests
from pyppeteer import page

class Crawler:
    def __init__(self, connection, options):
        self.connection = connection
        self.options = options
        self.result = None

    async def crawl(self, p : page):
        """ Crawls the given page object and returns an object describing the parsed results.
        :param p: the page that should be crawled
        """
        try:
            r = requests.post(self.connection, json={'url' : p.url, 'html' : await p.content()})
            crawl_result = json.loads(r.text)
            r.raise_for_status()
        except requests.exceptions.HTTPError as httperr:
            print("HTTP Error: ", httperr)
            return crawl_result # Return error loaded json
        except requests.exceptions.Timeout as timeouterr:
            return {'error':2, 'message': timeouterr}
        except requests.exceptions.RequestException as fatal:
            print(fatal)
            sys.exit(1)

        # Category needs to be retrieved seperatly
        crawl_result.update({'Category': await self.crawlCategory(p)})

        # Try to check the quality of the crawl.
        crawl_result.update({'content': await self.__check_crawl(p, crawl_result)})
        self.result = self.__parse(crawl_result)

    async def crawlCategory(self, p : page):
        pass # @TODO Retrieve the category. Consider using machine learning trained on https://data.world/crowdflower/url-categorization?

    def getResult(self):
        return self.result

    def __parse(self, answer):
        """ Takes an parsed object and removes/adds field so that it matches the required specification. """
        to_clean = ('lead_image_url', 'dek', 'next_page_url'
                    , 'domain', 'word_count', 'direction', 'total_pages'
                    , 'rendered_pages') # These are not needed? @TODO discuss with group!
        additional = {'Created': None, 'IsFromFeed': False, 'FeedUrl': None}
        for k in to_clean:
            answer.pop(k, None)
        answer.update(additional)

        return answer

    async def __check_crawl(self, p : page, crawl_result):
        """ Tries to assert on the quality of the crawl. It is possible that the mercury api was unsuccesful.
        If that is the case, we should try crawling manually again.
        :param p: the page that might have to be re-crawled
        :param crawl_result: the result of the first crawl
        """
        inner_text = await p.evaluate('document.body.innerText') # @TODO Improve using Query selectors

        #test = crawl_result.get('content', '') in inner_text and crawl_result.total_pages == crawl_result.rendered_pages
        test = crawl_result.get('total_pages', 1) == crawl_result.get('rendered_pages', 1)
        return crawl_result.get('content', '')if test else inner_text

        #@TODO