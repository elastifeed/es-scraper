import asyncio
import json
from pyppeteer import launch
from crawler import Crawler
from renderer import *

async def main():
    connection = "http://localhost:8080/parse"
    crawler = Crawler(connection, None)
    browser = await launch()
    page = await browser.newPage()
    test_url="https://news.yahoo.com/kremlin-says-north-korean-leader-kim-meet-putin-001907238.html"

    page.on("dialog", lambda dialog: asyncio.ensure_future(dismiss_dialog(dialog)))

    # Use GoogleBot UserAgent to handle redirections
    await page.setUserAgent("Googlebot/2.1 (+http://www.google.com/bot.html)")
    # Wait until domcontentloaded event is fired, otherwise, parsing results may be inconsitent
    await page.goto(test_url, {'waitUntil' : 'domcontentloaded'})

    await asyncio.gather(screenshot(page, "test.png"), renderPdf(page, "test.pdf")
                         , crawler.crawl(page))
    print(json.dumps(crawler.getResult()))

    await browser.close()

async def dismiss_dialog(dialog):
    """ Get rid of any obnoxious pop ups etc. """
    await dialog.dismiss

asyncio.get_event_loop().run_until_complete(main())
