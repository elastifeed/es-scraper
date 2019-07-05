#!/bin/bash


adresses="
https://www.golem.de/news/auslagerung-us-chiphersteller-umgehen-trumps-embargo-gegen-huawei-1906-142170.html
https://www.focus.de/politik/deutschland/rechtsextremen-auf-der-spur-stephan-e-gesteht-mord-an-luebcke-die-schluesselfrage-jetzt-war-er-allein_id_10867340.html
https://de.yahoo.com/sports/news/cathy-hummels-erst-dann-dortmund-203702414.html
https://news.yahoo.com/u-e-splits-u-over-112306909.html
https://www.huffpost.com/entry/elizabeth-warren-bernie-sanders-progressive-ideas-democratic-primary-debate_n_5d13f737e4b0e45560372da1
https://www.heise.de/ct/artikel/Raspberry-Pi-4-4-GByte-RAM-4K-USB-3-0-und-mehr-Rechenpower-4452964.html
https://www.dw.com/en/mediterranean-rescue-ship-brings-migrants-to-italy-defying-salvini/a-49363479
https://www.spiegel.de/lebenundlernen/schule/voelkerball-sollte-man-das-spiel-aus-der-schule-verbannen-a-1274627.html
https://www.engadget.com/2019/06/27/philips-hue-bluetooth-smart-light-bulbs/?guccounter=1&guce_referrer=aHR0cHM6Ly9uZXdzLmdvb2dsZS5jb20v&guce_referrer_sig=AQAAAHIxvUsf7WzEIyNF-bXEaFLm3rLG9MvRr9QsTlXNPWDWLgMKNwktCgpUJgQv8A9qkAhuG7TM92QCg7WzzSRtQcN9iJMT11f9CNEYF_cGiPzkIhKjPugdjES6OgKvLKKowuabY4MORsQgPO4PoHt40p_yJYWam7fB7GtOUnmvo4uq
https://www.chip.de/news/Wie-unter-macOS-Windows-10-Suche-runderneuert_170384191.html
https://www.pcwelt.de/news/US-Meteorologen-bitten-um-5G-Aufschub-10617010.html
"

i=0
for a in $adresses
do
    data="{\"url\":\""$a"\"}"
    echo "Sending: " $data
    curl -X POST -H "Content-Type: application/json" -d $data http://localhost:9090/scrape/all && echo &
    let i++
done

wait
