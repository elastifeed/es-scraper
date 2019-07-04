# es-scraper

[![Docker Repository on Quay](https://quay.io/repository/elastifeed/es-scraper/status "Docker Repository on Quay")](https://quay.io/repository/elastifeed/es-scraper)

Retrieves content from abitrary websites and fills it into the provided JSON interface. Additionally, it can render any website to a pdf and thumbnail.

## Dependencies
- `golang`
- Packages defined in go.mod

## Parameters
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_REGION
- AWS_ENDPOINT
- S3_ENDPOINT
- S3_BUCKET_NAME
- API_BIND
- MERCURY_URL	



## Endpoints
All endpoints expect a HTTP POST request with a URL to parse.
`{"url" : "<URL>"}`

 #### `/scrape/content `
 Performs a content scrape, using the [es-extractor](https://github.com/elastifeed/es-extractor). Also downloads the thumbnail and saves it to S3.
 
 Returns JSON with the following fiels:
 -  *author* - The author of the page. Can be null.
 -  *title* - The title (caption) of the url. Can be null.
 -  *date_published* - The publication date of the url. Can be null.
 -  *dek* - Can be null.
 -  *direction* - The reading direction of the content on the page. Can be null.
 -  *url* - The request url
 -  *excerpt* - "A small excerpt, most commonly the abstract of the article or the first few lines. Can be null.
 -  *raw_content* - Unformatted content retrieved from the url.
 -  *thumbnail* - Link to the page thumbnail in the storage. Can be empty if no thumbnail could be retrieved.
 -  *markdown_content* - The retrieved content formatted as markdown
 -  *total_pages* - The number of pages that were found to be part of this url
 -  *next_page_url* - Link to the next page. Null if the url only had one page
 -  *rendered_pages* - The number of pages that were parsed as part of this url.
 -  *word_count* - Counted words of the content

#### `/scrape/screenshot `
Takes a screenshot of the first page of the url. Saves this screenshot to S3.

Returns JSON with the following fields: 

-  *screenshot* - Link to the screenshot in the storage.

 
  #### `/scrape/pdf `
  Renders the page as a pdf file and saves it.
  
Returns JSON with the following fields: 

-  *pdf* - Link to the pdf in the storage.
 
  #### `/scrape/all `
  Performs all of the above actions combined. (Content with thumbnail, screenshot and pdf)

Returns a JSON contaiing all of the above fields.
