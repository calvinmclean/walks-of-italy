Goals:
  - Alert when tours open up on specific dates
  - Learn how far in the future each tour opens up (generally)
  - See when new dates are dropped and how many drop at once

Example `curl`:
```shell
curl 'https://api.ventrata.com/octo/availability' \
  -X 'POST' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -H 'Octo-Capabilities: octo/content,octo/pricing,octo/pickups,octo/extras,octo/offers,octo/resources' \
  -H 'Octo-Env: live' \
  -d '{"productId":"e9d2d819-5f04-4b1f-a07f-612387494b8f","optionId":"DEFAULT","localDateStart":"2025-09-01","localDateEnd":"2025-09-30","currency":"USD"}'
```

```shell
curl localhost:7077/tours -H "Content-Type: application/json" -X POST -d '{"Name": "VIP Vatican Key Master\'s Tour: Unlock the Sistine Chapel","URL": "https://www.walksofitaly.com/vatican-tours/key-masters-tour-sistine-chapel-vatican-museums/","ProductID": "e9d2d819-5f04-4b1f-a07f-612387494b8f"}'

curl localhost:7077/tours -H "Content-Type: application/json" -X POST -d '{"Name": "Private Vatican Tour: Vatican Museums, Sistine Chapel & St. Peter\'s","URL": "https://www.walksofitaly.com/vatican-tours/private-vatican-tour/","ProductID": "c40d8e0e-6756-463b-a052-982c77a707aa"}'

curl localhost:7077/tours -H "Content-Type: application/json" -X POST -d '{"Name": "The Complete Vatican Tour with Vatican Museums, Sistine Chapel & St. Peter\'s Basilica","URL": "https://www.walksofitaly.com/vatican-tours/complete-vatican-tour/","ProductID": "3b263ef8-c280-49cc-a74f-ac95aa2f1b58"}'

curl localhost:7077/tours -H "Content-Type: application/json" -X POST -d '{"Name": "Alone in the Sistine Chapel: VIP Entry at the Vatican\'s Most Exclusive Hours","URL": "https://www.walksofitaly.com/vatican-tours/vatican-after-hours-tour/","ProductID": "8c14824f-905d-4273-8b83-10b567db6e55"}'

curl localhost:7077/tours -H "Content-Type: application/json" -X POST -d '{"Name": "Pristine Sistine Early Entrance Small Group Vatican Tour","URL": "https://www.walksofitaly.com/vatican-tours/pristine-sistine-chapel-tour/","ProductID": "a1249220-e5d8-4983-93b2-c31ddfb3ccb8"}'
```
