# es-scraper
Retrieves content from abitrary websites and fills it into the provided JSON interface. Additionally, it can render any website to a pdf and thumbnail.

## Dependencies
- `python3.6+`
- Install dependencies with `python3 -m pip install --user -r requirements.txt`
- `docker` (Or `node, npm`)

  On linux, executing might cause an error: `No usable sandbox!`. To solve [configure a sandbox](https://github.com/GoogleChrome/puppeteer/blob/master/docs/troubleshooting.md#setting-up-chrome-linux-sandbox):
  ```
    # cd to the downloaded instance
    cd ~/.local/share/pyppeteer/local-chromium/575458/chrome-linux/
    sudo chown root:root chrome_sandbox
    sudo chmod 4755 chrome_sandbox
    # copy sandbox executable to a shared location
    sudo cp -p chrome_sandbox /usr/local/sbin/chrome-devel-sandbox
    # export CHROME_DEVEL_SANDBOX env variable
    export CHROME_DEVEL_SANDBOX=/usr/local/sbin/chrome-devel-sandbox
  ```
  or
  ```bash
      sudo sysctl -w kernel.unprivileged_userns_clone=1
  ```
