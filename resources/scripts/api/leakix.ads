-- Copyright © by Jeff Foley 2017-2023. All rights reserved.
-- Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.
-- SPDX-License-Identifier: Apache-2.0

local json = require("json")

name = "LeakIX"
type = "api"

function start()
    set_rate_limit(2)
end

function check()
    local c
    local cfg = datasrc_config()
    if (cfg ~= nil) then
        c = cfg.credentials
    end

    if (c ~= nil and c.key ~= nil and c.key ~= "") then
        return true
    end
    return false
end

function vertical(ctx, domain)
    local c
    local cfg = datasrc_config()
    if (cfg ~= nil) then
        c = cfg.credentials
    end

    if (c == nil or c.key == nil or c.key == "") then
        return
    end

    local resp, err = request(ctx, {
        ['url']=vert_url(domain),
        ['header']={
            ['api-key']=c.key,
            ['Accept']="application/json",
        },
    })
    if (err ~= nil and err ~= "") then
        log(ctx, "vertical request to service failed: " .. err)
        return
    elseif (resp.status_code < 200 or resp.status_code >= 400) then
        log(ctx, "vertical request to service returned with status: " .. resp.status)
        return
    end

    local d = json.decode(resp.body)
    if (d == nil) then
        log(ctx, "failed to decode the JSON response")
        return
    elseif (d.nodes == nil or #(d.nodes) == 0) then
        return
    end

    for _, node in pairs(d.nodes) do
        if (node ~= nil and node.fqdn ~= nil and node.fqdn ~= "") then
            new_name(ctx, node.fqdn)
        end
    end
end

function vert_url(domain)
    return "https://leakix.net/api/graph/hostname/" .. domain .. "?v%5B%5D=hostname&d=auto&l=1..5&f=3d-force"
end
