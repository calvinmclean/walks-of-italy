# Walks of Italy

This application allows tracking tours provided by [Walks of Italy](https://www.walksofitaly.com).

You can use the functions provided here to check specific date ranges for a tour's availability or run the server to get alerted when new tour dates are published.

Additionally, you can use the integrated AI tools and chat command with local Ollama models to interact with tour descriptions and availability.

## How To

This application uses the Ventrata API for Walks of Italy tours. It requires a token to get this data. You can get this token by accessing a tour page and searching `api.ventrata.com/octo/availability` in the Network tab of developer console. The token is shown in the `Authorization` header of requests.

If you want to use the AI chat with access to the descriptions API, you also need the token to access `tour-api.walks.org` which can be founds similarly to the previous example.

Now that you have a token, you can use the CLI:

```shell
export VENTRATA_TOKEN=token

go run cmd/walks-of-italy/main.go \
  search \
  --tour-id e9d2d819-5f04-4b1f-a07f-612387494b8f \
  --start 2025-11-15 --end 2025-12-20
```

### Load Data

You can load a few example tours into the DB with this command:

```shell
go run cmd/walks-of-italy/main.go \
  --db walks-of-italy.db \
  load \
  --data example-data.json
```

Or, once the server is running, use the API to insert:

```shell
curl localhost:7077/tours -H "Content-Type: application/json" -X POST -d '{"Name": "VIP Vatican Key Master\'s Tour: Unlock the Sistine Chapel","Link": "https://www.walksofitaly.com/vatican-tours/key-masters-tour-sistine-chapel-vatican-museums/","ProductID": "e9d2d819-5f04-4b1f-a07f-612387494b8f", "ApiUrl": "https://tour-api.walks.org/sites/walksofitaly/tour/key-masters-tour-sistine-chapel-vatican-museums"}'
```

### Run Server

```shell
export VENTRATA_TOKEN=token

go run cmd/walks-of-italy/main.go \
  --db walks-of-italy.db \
  serve
```

Then, visit http://localhost:7077 to see the UI!

### Use AI Chat

With Ollama running locally, you can chat about the tours you have in the DB:

```shell
export VENTRATA_TOKEN=token

go run cmd/walks-of-italy/main.go \
  --db server-backup.db \
  chat \
  --model qwen2.5:7b
```
