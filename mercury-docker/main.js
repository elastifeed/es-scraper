const Mercury = require('@postlight/mercury-parser');
const express = require('express');
const header = {'User-Agent':'Googlebot/2.1 (+http://www.google.com/bot.html)'}; // Always set User Agent in Header

var app = express(); // Use express to set up a basic server
app.use(express.json({limit : '100mb'}));

// React to post on /parse/url
app.post("/parse/url", async function(req, res) {
    const data = req.body;
    console.log(data)

    if (!data.url) {
        return res.status(400).send({
            error: 1,
            message: "Invalid JSON object"
        });
    }

    console.log(`Parsing from url: ${data.url}`)
    return res.json(await Mercury.parse(data.url, {headers : header, contentType : 'markdown'}))

});

// React to post on /parse/html
app.post("/parse/html", async function(req, res) {
    const data = req.body;

    if (!data.url || !data.html) {
        return res.status(400).send({
            error: 1,
            message: "Invalid JSON object"
        });
    }
    console.log(`Parsing from html: ${data.url}`)
    res.json(await Mercury.parse(data.url, {headers : header, contentType : 'markdown', html : data.html}))
});

app.listen(8080);
