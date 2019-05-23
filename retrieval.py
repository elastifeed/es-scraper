import asyncio
from crawler import Crawler
from sanic.log import logger
from pyppeteer import launch
from pyppeteer.browser import Browser
from pyppeteer.page import Page
from pyppeteer.errors import NetworkError
from renderer import screenshot, renderPdf

class RetrievalException(Exception):
    pass

class Retrieval:
    browser: Browser

    @staticmethod
    async def initialize():
        Retrieval.browser = await launch()
        p = (await Retrieval.browser.pages())[0]
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
        p : Page = (await Retrieval.browser.pages())[0]

        if p.url is url:
            return p # Page with correct url already loaded
        else:
            # Wait until domcontentloaded event is fired, otherwise, parsing results may be inconsitent
            logger.info("GOTO {}:\n\tfrom {}\n\tto {}".format(p, p.url, url))
            await p.goto(url, {'waitUntil': 'domcontentloaded'})

            return p

    @staticmethod
    async def _dismiss_dialog(dialog):
        """ Get rid of any obnoxious pop ups etc. """
        await dialog.dismiss
        logger.info("Dissmissed {} dialog".format(dialog))

    @staticmethod
    async def get_content(url):
        connection = "http://localhost:8080/parse/html"  # @TODO Extract this to a router?
        try:  # Assert that we have the necessary information to proccess the request
            page = await Retrieval._load(url)
        except AttributeError as e:
            logger.exception(e)
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
    async def get_thumbnail(url, path):
        try:  # Assert that we have the necessary information to proccess the request
            page = await Retrieval._load(url)
            await screenshot(page, path)
        except AttributeError as e:
            logger.exception(e)
            raise RetrievalException(e)
        except NetworkError as e:
            logger.error(e)
            raise RetrievalException(e)

    @staticmethod
    async def get_pdf(url, path):
        try:  # Assert that we have the necessary information to proccess the request
            page = await Retrieval._load(url)
            await renderPdf(page, path)
        except AttributeError as e:
            logger.exception(e)
            raise RetrievalException(e)
        except NetworkError as e:
            logger.error(e)
            raise RetrievalException(e)
