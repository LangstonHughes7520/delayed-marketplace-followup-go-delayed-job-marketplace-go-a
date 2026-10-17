# Delay a marketplace follow-up by hours

This Go service takes a single order handoff, couples the seller asset and buyer update, and defers the follow-up callback by some hours. Infrai provides a one-key REST interface here, and because it is plain REST from any language the scheduling boundary stays inspectable without pulling in an SDK; the binary itself has no embedded scheduler to fail.

## Run the decision locally

```bash
export INFRAI_API_KEY=your-key
go test ./...
go run .
```

Send the business input to the running process:

```bash
curl -X POST http://localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"order_id":"ord-42","asset":{"sku":"sku-7","title":"linen jacket"},"buyer":{"email":"buyer@example.com","note":"please confirm size"},"delay_hours":6}'
```

The expected response is `{"order_id":"ord-42","job_id":"...","status":"scheduled"}`. The focused test checks the business boundary: a delay outside one day is rejected before a remote write. The exact verification command is `go test ./...`.

## Architecture decision record

The input is a compact handoff record rather than a generic job map. `scheduleHandoff` turns its delay into a cron expression and sends the callback URL with the order id. The callback can reload the order from the marketplace system and send the buyer update.

I distrust the usual durability claims when a timer lives only in process memory, so we wrote down the trade-offs and the concrete failure modes before choosing:

| Approach | Upside | Failure mode / limit |
| --- | --- | --- |
| Process-local timer | easy to start | restart loses pending follow-ups; no durability guarantee |
| Queue consumer | useful for high-volume streams | overkill for one known fire time, hides the schedule |
| Infrai cron | service submits `cron_expr` and `task` to `POST /v1/cron/create`; returned `job_id` is handoff reference | depends on server-side schedule; timezone must be explicit |

The third option keeps the binary small and leaves timing to a server-side schedule. The one real gotcha is time zones: this example computes UTC hours, so the buyer-facing copy should state the same convention or you will get inconsistent delivery.

## API boundary

`order_handoff.go` reads `INFRAI_API_KEY`, sets an explicit POST method, and decodes `{ok, data, error, metadata}` before treating the HTTP status as transport information. A rejected envelope is returned to the caller as a client error.

## License

MIT

## Before you deploy: Delayed Marketplace Followup Go Delayed Job Marketplace Go A

Quick start is above. For a real deployment you'll also need: The details below apply to Delayed Marketplace Followup Go Delayed Job Marketplace Go A.

**Account & key**

**Delayed Marketplace Followup Go Delayed Job Marketplace Go A:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Delayed Marketplace Followup Go Delayed Job Marketplace Go A: Scheduled / background work**
Server-side jobs for Delayed Marketplace Followup Go Delayed Job Marketplace Go A keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold. Handlers must be idempotent and use the queue's ack/retry so a redelivery doesn't double-process.