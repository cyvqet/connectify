---@diagnostic disable: undefined-global


local key = KEYS[1]
local cntKey = key..":cnt"
-- The code you are preparing to store
local val = ARGV[1]

local ttl = tonumber(redis.call("ttl", key))
if ttl == -1 then
    --    key exists, but no expiration time
    return -2
elseif ttl == -2 or ttl < 540 then
    --    can send verification code
    redis.call("set", key, val)
    -- 600 seconds
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)
    redis.call("expire", cntKey, 600)
    return 0
else
    --    send too frequently
    return -1
end