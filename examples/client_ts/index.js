#!/usr/bin/env node

const util = require('util')

const { WidgetsClient } = require("../../gen/js/widgets/v0/widgets_pb_service")
const { GetRequest }    = require("../../gen/js/widgets/v0/widgets_pb")

const client = new WidgetsClient("http://localhost:80")

const req = new GetRequest()
req.setName("users/foo/widgets/blue")

client.get(req, (err, user) => {
    log(user)
})

function log(o) {
    console.log(util.inspect(o.toObject(), false, null, true))
}
