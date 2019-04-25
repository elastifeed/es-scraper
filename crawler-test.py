import asyncio
import json
from pyppeteer import launch

async def main():
    browser = await launch()
    page = await browser.newPage()
    test_url="https://news.yahoo.com/kremlin-says-north-korean-leader-kim-meet-putin-001907238.html"
#    await page.setRequestInterception(True)

    page.on("dialog", lambda dialog: asyncio.ensure_future(dismiss_dialog(dialog)))
    page.on("request", lambda request: asyncio.ensure_future(recieved_request(page, request)))
    # Use GoogleBot UserAgent to handle redirections
    await page.setUserAgent("Googlebot/2.1 (+http://www.google.com/bot.html)")
    await page.goto(test_url)

    content = await page.evaluate('document.body.innerText') # @todo Improve using Query selectors
    print(content)

async def recieved_request(page, request):
    print("\n[Recieved Request:::]")
    print(request.url)
    print(request.isNavigationRequest())
    print(request.redirectChain)

    if (request.isNavigationRequest() and request.redirectChain):
   #     request.abort()
        print("NO!")
        #if ("consent" in request.url):
        #    page.tap(page.querySelector('.consent-form'))

async def dismiss_dialog(dialog):
    """ Get rid of any obnoxious pop ups etc. """
    await dialog.dismiss

asyncio.get_event_loop().run_until_complete(main())
