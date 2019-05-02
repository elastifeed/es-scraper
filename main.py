import asyncio
import json
from pyppeteer import launch
from crawler import Crawler
from renderer import *

async def main():
    connection = "http://localhost:8080/parse/html"
    crawler = Crawler(connection, None)
    browser = await launch()
    page = await browser.newPage()
    test_url="https://www.golem.de/news/precision-dell-bringt-guenstige-whiskey-lake-workstations-mit-ubuntu-1905-140989.html"

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

asyncio.get_event_loop().run_until_complete(main())
