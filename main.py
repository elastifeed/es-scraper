import asyncio
import json
import sys
from pyppeteer import launch
from pyppeteer.browser import Browser
from pyppeteer.page import Page
from pyppeteer.errors import NetworkError
from sanic import Sanic
from sanic.log import logger
from sanic.response import json as json_resp
from sanic.response import text as text_resp
from crawler import Crawler
from renderer import *

app = Sanic(__name__)
browser: Browser
loop = asyncio.get_event_loop()

# TODO Save results of most recent call, being able to provide it on GET. Overwrite it on new POST


@app.route("/scrape", methods=['POST'])
async def scrape(request):
    """ Retrives content, renders a pdf and makes a thumbnail for the page.
    The request data must include:
        url : <page_url>, 
        thumbnail_path : <Full path to save the thumbnail>,
        pdf_path : <Full path to save the pdf>
    """
    logger.info(request.endpoint + " : " + str(request.json))

    #    test_url = 'https://www.golem.de/news/logistik-amazon-nutzt-in-europa-tagsueber-frachtflugzeuge-1905-141293.html'

   # await asyncio.gather(screenshot(page, "test.png"), renderPdf(page, "test.pdf")
   #                      , crawler.crawl(page))


    return  # TODO Full content


@app.route("/scrape/content", methods=['POST'])
async def content(request):
    """Endpoint to retrive the text content of a page. Requires a url."""
    logger.info(request.endpoint + " : " + str(request.json))
    connection = "http://localhost:8080/parse/html"  # @TODO Extract this to a router?
    try:  # Assert that we have the necessary information to proccess the request
        page = await load_page_for_url(request.json['url'])
    except AttributeError as att:
        logger.exception(att)
        return json_resp({'error': 1, 'message': str(att)}, status=400)
    except NetworkError as neterr:
        logger.error(neterr)
        return json_resp({'error': 1, 'message': str(neterr)+":"+request.json[ 'url' ]}, status=400)

    # Get a crawler object and retrieve the content
    crawler = Crawler(connection, None)
    await crawler.crawl(page)

    #print(json.dumps(crawler.getResult(), sort_keys=True, indent=4))
    return json_resp(crawler.getResult())


@app.route("/scrape/thumbnail", methods=['POST'])
async def thumbnail(request):
    """Endpoint to make a thumbnail of a page. Requires a path to save the thumbnail and a url."""
    logger.info(request.endpoint + " : " + str(request.json))
    try:  # Assert that we have the necessary information to proccess the request
        page = await load_page_for_url(request.json['url'])
        screenshot(page, request.json[ 'thumbnail_path' ])
    except AttributeError as att:
        logger.exception(att)
        return json_resp({'error': 1, 'message': str(att)}, status=400)
    except NetworkError as neterr:
        logger.error(neterr)
        return json_resp({'error': 1, 'message': str(neterr)+":"+request.json[ 'url' ]}, status=400)

    return text_resp(request.json.thumbnail_path)


@app.route("/scrape/pdf", methods=['POST'])
async def render(request):
    """Endpoint to render a pdf from a page. Requires a path to save said pdf and a url."""
    logger.info(request.endpoint + " : " + str(request.json))
    try:  # Assert that we have the necessary information to proccess the request
        page = await load_page_for_url(request.json[ 'url' ])
        renderPdf(page, request.json[ 'pdf_path' ])
    except AttributeError as att:
        logger.exception(att)
        return json_resp({'error': 1, 'message': str(att)}, status=400)
    except NetworkError as neterr:
        logger.error(neterr)
        return json_resp({'error': 1, 'message': str(neterr)+":"+request.json[ 'url' ]}, status=400)

    return text_resp(request.json.pdf_path)


async def dismiss_dialog(dialog):
    """ Get rid of any obnoxious pop ups etc. """
    await dialog.dismiss
    logger.info("Dissmissed {} dialog".format(dialog))


@app.listener("before_server_start")
async def initialize(app, loop):
    global browser
    browser = await launch()
    p : Page = (await browser.pages())[0]
    # Use GoogleBot UserAgent to handle redirections
    await p.setUserAgent("Googlebot/2.1 (+http://www.google.com/bot.html)")
    p.on("dialog", lambda dialog: asyncio.ensure_future(dismiss_dialog(dialog)))
     
    logger.debug("Launched browser.")


@app.listener("after_server_stop")
async def shutdown(app, loop):
    global browser
    await browser.close()
    logger.debug("Shutdown browser.")


async def load_page_for_url(url) -> Page:
    """ Gets the page that in the browser has loaded for this url.
    :param url: the url of the page
    """
    global browser
    p : Page = (await browser.pages())[0]

    if p.url is url:
        return p # Page with correct url already loaded
    else:
        # Wait until domcontentloaded event is fired, otherwise, parsing results may be inconsitent
        logger.info("GOTO {}:\n\tfrom {}\n\tto {}".format(p, p.url, url))
        await p.goto(url, {'waitUntil': 'domcontentloaded'})

        return p

if __name__ == "__main__":
    app.run('127.0.0.1', 8000)
