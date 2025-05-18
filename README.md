Goals:
  - Alert when tours open up on specific dates
  - Learn how far in the future each tour opens up (generally)
  - See when new dates are dropped and how many drop at once

Example `curl`:
```shell
curl 'https://api.ventrata.com/octo/availability' \
  -X 'POST' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer b082247d-90d9-4e44-8623-3a8a4b5c14de' \
  -H 'Octo-Capabilities: octo/content,octo/pricing,octo/pickups,octo/extras,octo/offers,octo/resources' \
  -H 'Octo-Env: live' \
  -d '{"productId":"e9d2d819-5f04-4b1f-a07f-612387494b8f","optionId":"DEFAULT","localDateStart":"2025-09-01","localDateEnd":"2025-09-30","currency":"USD"}'
```


```shell
curl localhost:7077/tours -H "Content-Type: application/json" -X POST -d '{"Name": "VIP Vatican Key Master\'s Tour: Unlock the Sistine Chapel","URL": "https://www.walksofitaly.com/vatican-tours/key-masters-tour-sistine-chapel-vatican-museums/","ProductID": "e9d2d819-5f04-4b1f-a07f-612387494b8f"}'
```
