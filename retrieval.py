import asyncio
import difflib
from uuid import uuid4
from crawler import Crawler
from pyppeteer import launch
from pyppeteer.browser import Browser
from pyppeteer.errors import NetworkError
from pyppeteer.page import Page
from renderer import screenshot, renderPdf
from sanic.log import logger

class RetrievalException(Exception):
    pass

class Retrieval:
    browser: Browser

    @staticmethod
    async def initialize():
        Retrieval.browser = await launch()
        p = await Retrieval._get_page()
        # Use GoogleBot UserAgent to handle redirections
        await p.setUserAgent("Googlebot/2.1 (+http://www.google.com/bot.html)")
        p.on("dialog", lambda dialog: asyncio.ensure_future(Retrieval._dismiss_dialog(dialog)))
        
        logger.debug("Launched browser.")

    @staticmethod
    async def shutdown():
        await Retrieval.browser.close()
        logger.debug("Shutdown browser.")

    @staticmethod
    async def _load(url):
        """ Gets the page that in the browser has loaded for this url.
        :param url: the url of the page
        """

        def __url_compare(expected, actual):
            """Compares two urls and determines whether they are the same. The need for this is because for example http://google.de and https://www.google.de/ are the same adress and thus should'nt be loaded twice
            """
            allowed = ['w', 's', '/', '.']
            diff = [change for change in difflib.ndiff(actual, expected) if change[0] != ' ']
            for i in diff:
                if i[-1] not in allowed: # Checking the last char, if it isn't part of the exclusion list, we can't guarantee that the urls are the same.
                    return False
            return True
            
        p : Page = await Retrieval._get_page()

        if __url_compare(p.url, url):
            return p # Page with correct url already loaded
        else:
            # Wait until domcontentloaded event is fired, otherwise, parsing results may be inconsitent
            logger.info("GOTO {}:\n\tfrom {}\n\tto {}".format(p, p.url, url))
            await p.goto(url, {'waitUntil': 'domcontentloaded'})

            return p

    @staticmethod
    async def _get_page():
        """Retrieves an already initialized page object of the browser.
        :raises RetrievalException: If the browser has no pages."""
        for target in Retrieval.browser.targets():
            if target.type == "page":
                return await target.page()
        # No page found
        raise RetrievalException('Browser contains no pages.')

    @staticmethod
    async def _dismiss_dialog(dialog):
        """ Get rid of any obnoxious pop ups etc. """
        await dialog.dismiss
        logger.info("Dissmissed {} dialog".format(dialog))

    @staticmethod
    async def get_content(url):
        connection = "http://localhost:8080/mercury/html"  # @TODO Extract this to a router?
        try:  # Assert that we have the necessary information to proccess the request
            page = await Retrieval._load(url)
        except AttributeError as e:
            logger.warning(e, exc_info=True)
            raise RetrievalException(e)
        except NetworkError as e:
            logger.error(e)
            raise RetrievalException(e)

        # Get a crawler object and retrieve the content
        crawler = Crawler(connection, None)
        await crawler.crawl(page)

        #print(json.dumps(crawler.getResult(), sort_keys=True, indent=4))
        return crawler.getResult()

    @staticmethod
    async def get_thumbnail(url, path=None):
        path = path or f"/tmp/{uuid4()}.png"
        try:  # Assert that we have the necessary information to proccess the request
            page = await Retrieval._load(url)
            await screenshot(page, path)
            return path
        except AttributeError as e:
            logger.warning(e, exc_info=True)
            raise RetrievalException(e)
        except NetworkError as e:
            logger.error(e)
            raise RetrievalException(e)

    @staticmethod
    async def get_pdf(url, path=None):
        path = path or f"/tmp/{uuid4()}.pdf"
        try:  # Assert that we have the necessary information to proccess the request
            page = await Retrieval._load(url)
            await renderPdf(page, path)
            return path
        except AttributeError as e:
            logger.warning(e, exc_info=True)
            raise RetrievalException(e)
        except NetworkError as e:
            logger.error(e)
            raise RetrievalException(e)
