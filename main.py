import asyncio
import json
import sys
from pyppeteer import launch
from flask import Flask
from crawler import Crawler
from renderer import *
app = Flask(__name__)
loop = asyncio.get_event_loop ()

@app.route("/scrape")
def notify():
    loop.run_until_complete(scrape())

async def scrape():
    connection = "http://localhost:8080/parse/html"
    #test_url=sys.argv[1]
    test_url = 'https://www.cnet.com/news/uber-drivers-demand-fair-pay-ahead-of-companys-multibillion-dollar-ipo/'

    crawler = Crawler(connection, None)
    browser = await launch()
    page = await browser.newPage()

    page.on("dialog", lambda dialog: asyncio.ensure_future(dismiss_dialog(dialog)))

    # Use GoogleBot UserAgent to handle redirections
    await page.setUserAgent("Googlebot/2.1 (+http://www.google.com/bot.html)")
    # Wait until domcontentloaded event is fired, otherwise, parsing results may be inconsitent

    await page.goto(test_url, {'waitUntil' : 'domcontentloaded'})

    await asyncio.gather(screenshot(page, "test.png"), renderPdf(page, "test.pdf")
                         , crawler.crawl(page))
    print(json.dumps(crawler.getResult(), sort_keys=True, indent=4))

    await browser.close()

async def dismiss_dialog(dialog):
    """ Get rid of any obnoxious pop ups etc. """
    await dialog.dismiss

if __name__ == "__main__":
    app.run()
