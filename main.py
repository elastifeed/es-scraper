import asyncio
from sanic import Sanic
from sanic.log import logger
from sanic.response import json as json_resp
from sanic.exceptions import abort
from sanic.response import text as text_resp
from retrieval import Retrieval, RetrievalException

app = Sanic(__name__)
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

    try:
        url = request.json['url']
    except KeyError:
        abort(500)
    )

 await asyncio.gather(screenshot(page, "test.png"), renderPdf(page, "test.pdf")
                        , crawler.crawl(page))


    return  # TODO Full content


@app.route("/scrape/content", methods=['POST'])
async def content(request):
    """Endpoint to retrive the text content of a page. Requires a url."""
    logger.info(request.endpoint + " : " + str(request.json))
    try:
        return json_resp(await Retrieval.get_content(request.json["url"]))
    except ( RetrievalException, KeyError ):
        return abort(500)


@app.route("/scrape/thumbnail", methods=['POST'])
async def thumbnail(request):
    """Endpoint to make a thumbnail of a page. Requires a path to save the thumbnail and a url."""
    logger.info(request.endpoint + " : " + str(request.json))
    try:
        return text_resp(
            await Retrieval.get_thumbnail(
                request.json["url"],
                f"/tmp/test.png"
                )
            )
    except ( RetrievalException, KeyError):
        return abort(500)


@app.route("/scrape/pdf", methods=['POST'])
async def render(request):
    """Endpoint to render a pdf from a page. Requires a path to save said pdf and a url."""
    try:
        return text_resp(
            await Retrieval.get_pdf(
                request.json["url"],
                f"/tmp/test.pdf"
                )
            )
    except ( RetrievalException, KeyError):
        return abort(500)


@app.listener("before_server_start")
async def initialize(app, loop):
    await Retrieval.initialize()


@app.listener("after_server_stop")
async def shutdown(app, loop):
    await Retrieval.shutdown()
    logger.debug("Shutdown browser.")


if __name__ == "__main__":
    app.run('127.0.0.1', 8000)
