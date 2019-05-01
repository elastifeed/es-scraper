import asyncio
from pyppeteer import page

async def screenshot(p : page, savePath):
    await p.setViewport({'width':700, 'height':900, 'isMobile':True})
    await p.screenshot(path=savePath, clip={'x':0, 'y':0, 'width':700, 'height':900})

async def renderPdf(p : page, savePath):
    await p.emulateMedia('screen')
    await p.pdf(path=savePath)
