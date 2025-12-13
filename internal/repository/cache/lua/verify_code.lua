---@diagnostic disable: undefined-global

local key = KEYS[1]
local cntKey = key..":cnt"
-- The code the user entered
local expectedCode = ARGV[1]

local cnt = tonumber(redis.call("get", cntKey))
local code = redis.call("get", key)

if cnt == nil or cnt <= 0 then
--    verification count depleted
    return -1
end

if code == expectedCode then
    redis.call("set", cntKey, 0)
    return 0
else
    redis.call("decr", cntKey)
    --    not equal, the user entered the wrong code
    return -2
end