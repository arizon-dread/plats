# Plats
Plats is a swedish word that translates to location.
## Zipcodes and cities
The main/initial purpose of this API is to create a unified endpoint for a fan-out of an upstream collection of API's to ask for a city by using a zip-code, and to cache the response in a redis/valkey-like cache, to offload the upstream API's.  
## Endpoints
`GET /healthz` - API-endpoint for Kubernetes to see if the application is healthy. At the time of writing, this will return `200 healthy` as long as the app is running.  
`GET /metrics` - API-endpoint for Prometheus metrics scraping. Adds custom endpoints for all upstream API's and also one for the cache.  
`GET /v1/zip/${zip}` - API-endpoint that returns a plain text city name for the given zip code.  

## Config
sample config:
```yaml
cache:
  url: localhost
  #user:
  #pass:
  protocol: tcp
  port: 6379
apis: 
- name: "my_api" #use snake_case for prometheus' sake.
  url: "https://api.com/"
  path: "?query=${zip}&apikey=${apikey}"
  apiKey: #will be replaced by ${apis[#].name}_apikey from the env
  responseCityKey: results.0.city #this returns a more complex json structure.
  logHeaders: [
    MyQuotaResponseHeader, #will be logged as "<datestring> my_api.MyQuotaResponseHeader: <value>"
    MyOtherRespHeader
  ]
  fallback: true
- name: "external_api" #use snake_case for prometheus' sake.
  url: https://external-api.com
  path: /api/postal/${zip}
  responseCityKey: city #this API retuns a flat, small json that has one key/value-pair so the gjson mapping is simple.
```  
### Environment variables
`path=/go/bin/config` - will make the app look for a config file in the given path.  
`environment=development` - will make the app look for `development.yaml` as the config file, it will suffix the `environment` env var value with `.yaml`.  
`my_api_apikey=s3cr3t-4p1k3y` - will use the pattern `${apis[#].name}_apikey` in the env to replace each setting in the `apis` array of entries in the config object, which in turn will be used to replace the `${apikey}` value inside the url of the API.   

### Response json and responseCityKey
The [gjson](https://github.com/tidwall/gjson) project is used to map which json key in the response in the upstream API to use to get the value for the response to the downstream client.  
See the gjson documentation for details on how to drill down in a json document.

### LogHeaders
This []string will search for response headers in the upstream API responses and log them, if they're found. If you have an upstream API that responds with a quota on how many calls you have left, as a header on each response, you can specify the header and then monitor the log for a pattern using your favourite log muncher.

## Performance and cost 
The focus of the API is performance, the quickest response from upstream will be returned to the caller, and also cached. The cache will be checked before calling upstream API's.  
The API's can be graded into two categories, fallbacks or main API's. This way, you can gradually decrease the cost of hits on the upstream API's as you build the local cache, and also prioritise speed or low cost upstream API's in your main API's collection and only ask the costly or slow API's if you don't get a hit in your main ones.

## Statistics/metrics
The API will provide metrics on the prometheus format on the `/metrics` endpoint. It will dynamically create a counter for each API you add to the config list, and also a counter for the cache. You can then monitor, or create graphs that detail how the fan-out is working, which API's are fast, to what grade the cache is used etc. Please note that only the fastest upstream hit will generate an uptick, even if the other endpoints in your config are hit but return slower. That means you can't rely on the `/metrics` endpoint to warn you when you're getting near a rate limit for your upstream API's, some of the hits will be swallowed if there are other, faster API's in your list. You can, however, specify response headers for each API, that will be logged regardless if the hit is winning the speed contest towards the other upstream API's or not.    
