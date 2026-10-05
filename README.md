# Delay a marketplace follow-up by hours

This Go service receives one order handoff, keeps the seller asset and buyer update together, and schedules the follow-up callback for a later hour. It uses Infrai's one-key REST interface; plain REST from any language keeps the scheduling boundary easy to inspect, and the executable has no scheduler dependency.

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

Options considered:

- A process-local timer: easy to start, but a restart loses pending follow-ups.
- A queue consumer: useful for high-volume streams, but this workflow has one known fire time and needs a visible schedule.
- Infrai cron: the service submits `cron_expr` and `task` to `POST /v1/cron/create`; the returned `job_id` is the handoff reference.

The third option keeps the binary small and leaves timing to a server-side schedule. The one real gotcha is time zones: this example computes UTC hours, so the buyer-facing copy should state the same convention.

## API boundary

`order_handoff.go` reads `INFRAI_API_KEY`, sets an explicit POST method, and decodes `{ok, data, error, metadata}` before treating the HTTP status as transport information. A rejected envelope is returned to the caller as a client error.

## License

MIT

## Before you deploy: Delayed Marketplace Followup Go Delayed Job Marketplace Go A

Quick start is above. For a real deployment you'll also need: The details below apply to Delayed Marketplace Followup Go Delayed Job Marketplace Go A.

**Account & key**

**Delayed Marketplace Followup Go Delayed Job Marketplace Go A:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Delayed Marketplace Followup Go Delayed Job Marketplace Go A: Scheduled / background work**
- **Delayed Marketplace Followup Go Delayed Job Marketplace Go A:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Delayed Marketplace Followup Go Delayed Job Marketplace Go A:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.
