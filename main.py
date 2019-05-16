import asyncio
import json
import sys
from pyppeteer import launch
from sanic import  Sanic
from sanic.response import json as json_resp
from crawler import Crawler
from renderer import *

app = Sanic(__name__)
loop = asyncio.get_event_loop ()

@app.route("/scrape", methods=['POST'])
async def scrape(request):
    print(request)
    connection = "http://localhost:8080/parse/html"
    test_url = 'http://www.golem.de'

    crawler = Crawler(connection, None)
    page = await browser.newPage()

    page.on("dialog", lambda dialog: asyncio.ensure_future(dismiss_dialog(dialog)))

    # Use GoogleBot UserAgent to handle redirections
    await page.setUserAgent("Googlebot/2.1 (+http://www.google.com/bot.html)")

    # Wait until domcontentloaded event is fired, otherwise, parsing results may be inconsitent
    await page.goto(test_url, {'waitUntil' : 'domcontentloaded'})

   #print(await page.evaluate('document.body.innerText'))

    #await asyncio.gather(screenshot(page, "test.png"), renderPdf(page, "test.pdf")
    #                     , crawler.crawl(page))
    await crawler.crawl(page)
    print(json.dumps(crawler.getResult(), sort_keys=True, indent=4))


    return json_resp(crawler.getResult())

async def dismiss_dialog(dialog):
    """ Get rid of any obnoxious pop ups etc. """
    await dialog.dismiss

@app.listener("before_server_start")
async def initialize(app, loop):
    global browser
    browser = await launch()

@app.listener("after_server_stop")
async def shutdown(app, loop):
    await browser.close()

if __name__ == "__main__":
    app.run('127.0.0.1', 8000)
