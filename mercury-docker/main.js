const Mercury = require('@postlight/mercury-parser');
const express = require('express');
const header = {'User-Agent':'Googlebot/2.1 (+http://www.google.com/bot.html)'}; // Always set User Agent in Header
var app = express(); // Use express to set up a basic server

app.use(express.json({limit : '100mb'}));
// React to post on /parse
app.post("/parse", function(request, response) {
    data = request.body;

    if (data.url == null && data.html == null) {
        response.status(403);
        sendResult(response, requestError());
    } else if (data.html == null) { // No preloaded HTML specified - Parse from URL.
        console.log("Parsing from url:")
        console.log(data.url)
        Mercury.parse(data.url, {headers : header})
            .then(result => sendResult(response, result));
    } else {
        console.log("Parsing from html:")
        console.log(data.url)
        Mercury.parse(data.url, {headers : header, contentType : 'text', html : data.html})
           .then(result => sendResult(response, result));
    }
});

app.listen(8080);

function sendResult(response, data) {
    console.log("Sending: " + data)
    response.json(data); // Send the data
}

function requestError() {
    return {'error':1, 'message':'Invalid JSON object.'};
}