# M4 — relay / "from anywhere"

**Status:** planned, not started. M4 comes after M1–M3. v1 is LAN-only.

**Goal:** add a relay so the Android app can reach the host across NAT / from anywhere, without a directly reachable IP, gated by consent.

**Out of scope for v1:** everything in M4. v1 is LAN control only.

**Depends on:** M1–M3 being done, or at least M1–M3's API/transport surfaces being defined well enough to front with a relay.

---

## M4.1 — relay design decision

**Goal:** decide the relay shape before building it.

**Acceptance criteria:**

- [ ] The plan records whether the relay is a Cloudflare Worker-based relay, a separate self-hosted relay, or another shape.
- [ ] The plan records the pairing model: room-secret handshake, Noise-encrypted transport, consent plugin analog for host egress, or another model.
- [ ] The plan records what "from anywhere" means for v1 vs. later (for example: relay reaches the host across NAT but still LAN-speed expectations, vs. a full internet-path relay with push).

**Test cases:**

- [ ] N/A for now; this task is a design decision recorded in the plan.

**Out of scope:**

- [ ] Implementation. This task is the design record.

**Depends on:** M1–M3 for the surfaces the relay fronts.

---

## M4.2 — relay implementation (placeholder)

**Goal:** implement the relay transport and pairing so the Android app can reach the host from anywhere.

**Acceptance criteria:**

- [ ] To be written when M4 is scheduled.

**Test cases:**

- [ ] To be written when M4 is scheduled.

**Depends on:** M4.1.

---

## M4 done-ness

To be defined when M4 is scheduled. v1 explicitly excludes this.
