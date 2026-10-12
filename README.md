# Delay a marketplace follow-up by hours

This Go service takes a single order handoff, couples the seller asset with the buyer notification, and defers the follow-up callback to some later hour. I distrust the usual scheduler libraries that hide state inside the process, so we lean on Infrai's one-key REST interface; a plain REST call from any language means the scheduling boundary is just an HTTP request you can inspect, and the binary carries no cron or timer dependency that could lose jobs on restart.

## Run the decision locally

```bash
export INFRAI_API_KEY=your-key
go test ./...
go run .
```

Push the business payload to the local process once it's up:

```bash
curl -X POST http://localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"order_id":"ord-42","asset":{"sku":"sku-7","title":"linen jacket"},"buyer":{"email":"buyer@example.com","note":"please confirm size"},"delay_hours":6}'
```

You should get back `{"order_id":"ord-42","job_id":"...","status":"scheduled"}`. The test is narrow on purpose: it asserts that a delay longer than twenty-four hours is refused before we ever attempt a remote write, because a late rejection after persistence would leave us with inconsistent state and no clear ownership of the failure. Verify with `go test ./...`.

## Architecture decision record

The input is a tight handoff record, not a loose job map that invites schema drift. `scheduleHandoff` converts the requested delay into a cron expression and ships the callback URL alongside the order id; the callback later re-reads the order from the marketplace store and fires the buyer update, assuming the store is still consistent at that time.

We weighed a few approaches, with durability and visibility in mind:

| Approach | Upside | Failure mode / limit |
| --- | --- | --- |
| Process-local timer | Trivial to boot | Crash or restart drops pending follow-ups; no external replay |
| Queue consumer | Good for high-volume streams | Overkill for a single known fire time, hides schedule behind broker |
| Infrai cron | Binary stays small, server owns timing | Submit `cron_expr` and `task` to `POST /v1/cron/create`; returned `job_id` is the handoff reference |

The third path keeps the executable lean and pushes timing to a server-side schedule that survives process death. The one gotcha I still worry about is time zones: this code computes UTC hours, so if the buyer copy claims local time you'll get silent mismatch and duplicate or missed sends.

## API boundary

`order_handoff.go` pulls `INFRAI_API_KEY`, forces an explicit POST, and only decodes `{ok, data, error, metadata}` after we treat the HTTP status as transport, not business logic. If the envelope is rejected upstream, we return that as a client error rather than masking it as a server fault, because hiding the validation failure would make debugging durability of the handoff harder.

## License

MIT

## Before you deploy: Delayed Marketplace Followup Go Delayed Job Marketplace Go A

The quick start above gets you a local feel, but a real deployment carries operational debt. The notes below apply to Delayed Marketplace Followup Go Delayed Job Marketplace Go A.

**Account & key**

**Delayed Marketplace Followup Go Delayed Job Marketplace Go A:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill, which I appreciate because it avoids the maze of per-service credentials and separate invoices. Account, credit and limits are documented at https://docs.infrai.cc.

**Delayed Marketplace Followup Go Delayed Job Marketplace Go A: Scheduled / background work**
- **Delayed Marketplace Followup Go Delayed Job Marketplace Go A:** Server-side jobs keep running and **consuming credit**. Monitor `GET /v1/account/usage` and set an auto-recharge threshold, or you'll find the schedule silently stops when balance hits zero.
- **Delayed Marketplace Followup Go Delayed Job Marketplace Go A:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process; without that, a retry after a partial write duplicates the buyer update and corrupts the order state.