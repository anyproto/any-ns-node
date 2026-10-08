# any-ns-node 0.7.2: names stay reserved to their identity

Status: approved design, 2026-10-08. Base: `v0.7.1` (`ded2b34`). Branch: `fix/ns-names-reserved`.

## Why

0.7.1 (GO-7567) made the name cache follow the chain, including **freeing** names that lapsed
(expired + 90 day grace): the record became a tombstone and `IsNameAvailable` reported the name
available. In production (deployed 2026-10-07) this caused:

1. **Data loss.** A tombstone is written with `ReplaceOne` of a bare observation
   (`cache/read.go` `observation()` + `cache/store.go` `applyObservationTx`): the owner, AnyID,
   space ID and expiry of the record are wiped. ~160 prod records by 2026-10-08 (the prod count grows until 0.7.2 is deployed) (owners recoverable
   from the 2026-10-07 nightly dump, `db3:/var/lib/backup/mongodb/ns-cache-20261007.jsonl`).
2. **A policy change nobody wanted.** Product decision (2026-10-08): **a name stays reserved to the
   identity that registered it, forever, even when it has lapsed on chain.** Moving a name to
   another identity is a support action.
3. **Provider quota exhaustion.** Lookups of expired/lapsed/tombstoned names queue chain re-reads,
   and the periodic scan re-reads every record at expiry, at lapse and then daily. Together with a
   `-refresh-cache` pass this exhausted the Infura daily credits (2026-10-07 18:00 - 03:00 MSK every
   chain read of both prod nodes failed with 429).

What 0.7.1 got right and 0.7.2 keeps: renewals reach the cache (`GetOperation` -> `RefreshAfterOperation`
for a cached name, +5/+30 min re-reads), GO-7482 (no false `Completed`, `PendingOrNotFound`), canonical
names + `canon` + unique `{name:1}` index, transactional block-ordered writes, replica set
requirement, ctx in contract calls, `-dedupe-cache`.

## Principles

- Registrations only happen through Anytype (the registrar controller is private): every on-chain
  change of a name is the result of one of our operations. **The cache is the record of who holds a
  name.** The chain is read to learn the result of our own operations, never to decide that a name
  is free.
- **Nothing is ever removed or wiped automatically.** Only an explicit operator action deletes a record.
- **Chain reads are driven by operations, not by lookups or time.** Expiry is a comparison of
  `name_expires` with the clock, it needs no chain read.

## Behaviour

### 1. A lapse keeps the record (`lapsed` flag)

- `readRegistration` (`cache/read.go`) distinguishes three outcomes instead of two:
  registered / lapsed / not registered.
  - registry owner present, registrar expiry present, `isLapsed` -> **lapsed** (with the expiry).
  - zero registry owner, registrar expiry present, `isLapsed` -> **lapsed**.
  - zero owner and no expiry -> not registered (as today).
  - everything else unchanged (owner + no expiry stays `errInconsistentRegistry`, a failure).
- `readNameData` on a lapsed name returns an observation with `NameExpires`, `RegistryOwner`,
  the block fields and `Lapsed = true`, **without** the enrichment reads (owner/AnyID/space: after a
  lapse the NameWrapper no longer reports the owner, the reads would fail forever). A sentinel
  (e.g. `errLapsed`) tells `refresh` which case it is.
- `refresh` on a lapsed observation:
  - stored record exists (removed or not): in one transaction, if the observation is newer
    (`newer()`), **update in place** (`$set`, never `ReplaceOne`): `name_expires`, `lapsed: true`,
    `registry_owner`, `observed_*`, `fork_read_at` rules as today, re-reads merged/done as today,
    `repair_at` recomputed. Owner fields, AnyID, space ID, `canon` are **kept**. `incomplete` /
    `refresh_needed` are cleared (nothing more to read). `removed` is left as it is (a 0.7.1
    tombstone has no owner fields; `-restore-tombstones` brings them back and clears it). Return
    the stored record, error `nil`.
  - no stored record: nothing is written, return `ErrNameNotRegistered` (the name was never ours,
    or the cache lost it: an operator matter, see `-refresh-cache` reporting).
- A later refresh that finds the name registered again (re-registration by the same owner, or a
  support transfer) replaces the record as today and clears `lapsed`.
- `NameDataItem` gets `Lapsed bool \`bson:"lapsed,omitempty"\``: "the chain lets this name lapse
  (past grace) at ObservedBlock; the cache keeps it reserved for its owner".

### 2. "Not registered" never removes

- Remove `confirmNotRegistered`, the finalized/safe block confirmation for removals, the
  `confirm` / `background`-reconfirm options that only served it, and every tombstone write.
- `refresh` when the latest block says not registered:
  - no stored record: return `ErrNameNotRegistered` (unchanged: `GetOperation` reports
    `PendingOrNotFound`, GO-7482).
  - stored record: keep it as it is, log a warning, return a distinct sentinel
    (e.g. `errNotOnChain`, wrapping `ErrNameNotRegistered` so existing callers keep their
    meaning) that `failed()` treats as **settled** (no backoff loop): right after a registration
    this is provider lag, which the scheduled +5/+30 re-reads pick up.
- `Removed` stays in `NameDataItem` only to read existing tombstones. **0.7.2 serves a record with
  `removed: true` as taken** (`IsNameAvailable` -> not available, with whatever owner fields it
  has; reverse lookups keep skipping it because its owner fields are empty anyway). It never
  writes `removed: true`. The alias check (`liveAlias`) counts removed records too.
- Remove `-purge-tombstones` (`PurgeTombstones`, `ErrIncompleteRecords`, `PurgeStats`, the CLI
  flag, docs). A rollback to 0.6.9 needs no preparation: 0.6.9 also serves any record as taken.

### 3. Background refresh only for operations and incomplete records

- `needsRefresh` (lookup-triggered): only `RefreshNeeded` records (incomplete enrichment, failed
  post-operation refresh), respecting `RefreshNextAt`. Expired, lapsed, grace and removed records
  are **never** handed to the background by a lookup.
- `repairAt`: due only for `RefreshNeeded` and pending `Rereads`. Remove `expiryCheckAt`,
  `expiredRepairInterval`, `expiredRefreshInterval` (keep a settle interval for aliases if
  `settleAlias` still needs one).
- `repairOnce`: records written by 0.7.1 carry `repair_at` values from the old rules (expiry,
  lapse, daily). Before refreshing a due record, recompute `repairAt` from its stored fields
  (project the fields needed); if it is not due under the new rules, `$set`/`$unset` its
  `repair_at` accordingly and skip it **without a chain read**. This keeps the first scans after
  the upgrade from re-reading hundreds of names.
- Failure backoff becomes exponential per record: a new field
  `refresh_failures int \`bson:"refresh_failures,omitempty"\`` incremented on each failed
  background refresh, reset on success; backoff = min(1 min * 2^(failures-1), 24 h), failures
  counted including this one. Applies to
  `backOff` / `retryAfterBackoff` and `markRefreshNeeded`. A successful write of the record
  (applyObservationTx replace, lapse update) resets it.
- Post-operation behaviour stays: `RefreshAfterOperation`, `UpdateInCacheAfterOperation`,
  re-reads at +5/+30 min, leases.

### 4. Server-side reservation on registration

- `AdminNameRegisterSigned` (`anynsrpc/anynsrpc.go`), after the parameter check and before
  `aa.AdminNameRegister`: look up the canonical name in the cache (any record, removed or not,
  aliases too: reuse `IsNameAvailable`). If a record exists and its `OwnerAnyAddress` is empty or
  differs from the request's `OwnerAnyAddress`, reject with a clear error
  (`ErrNameReserved`: "the name is reserved for another identity"). Same owner -> allowed (pp's
  renew-past-grace -> register fallback, a re-registration after a lapse).
- A cache read error rejects the request (fail closed); the payment node retries.
- `AdminNameRenewSigned`: no change.
- Not in scope: the user-operation path (`CreateUserOperation` / `GetDataNameRegister*`) builds
  calldata for client-signed registrations; add the same check there only if it is cheap and clearly
  correct, otherwise list it as a follow-up.

### 5. Support transfer: `-release-name`

- New maintenance flag `-release-name <name>`: normalizes the name, prints the record (owner
  fields, expiry, lapsed, removed), and with `-refresh-apply` deletes it (one `DeleteOne` by
  `_id`, logged at warn with the full record). Dry run by default. Aliases of the name are listed,
  not deleted. After a release the name can be registered for a new identity.

### 6. Migration: `-restore-tombstones -restore-from <file>`

- Input: a `mongoexport` JSON-lines file of the cache (MongoDB Extended JSON, e.g.
  `{"_id":{"$oid":...},"name":"x.any","owner_eth_address":"0x..","owner_scw_eth_address":"0x..",
  "owner_any_address":"A..","space_id":"","name_expires":{"$numberLong":"..."}|number, ...}`).
  Parse with `bson.UnmarshalExtJSON`.
- For each record with `removed: true` (one transaction each, filter `{_id, removed: true}`):
  - found in the file with a non-empty `owner_any_address`: `$set` owner fields, `space_id`,
    `name_expires` from the file, `lapsed: true` (if `isLapsed(name_expires, now)`, otherwise
    leave it unset and set `refresh_needed` so the background re-reads it once), `canon`
    (canonical of the name) if missing; `$unset` `removed`, `incomplete`; recompute `repair_at`.
    Keep the `observed_*` block fields of the tombstone.
  - not found / no owner: list it, leave it (served as taken without an owner).
- Dry run by default (`-refresh-apply` writes). Output:
  `restore DRY RUN|APPLIED: tombstones=N restored=N skipped=N missing=N conflicts=N owner-differs=N`
  + `  missing:`, `  conflict:` and `  owner-differs:` lines (see the README). Exit 1 if any of
  them is printed (operator attention), 0 otherwise.

### 7. `-refresh-cache` changes

- New flag `-rpc-url <url>`: maintenance runs only; overrides `contracts.gethUrl` for this process
  (e.g. Alchemy), so a full pass does not spend the live nodes' provider quota. Never logged in full
  (log the host only).
- Stats: drop `Removed`, `NotFinal`; add `Lapsed` (records marked/kept lapsed) and `NotOnChain`
  (stored record, chain says never registered: kept). Output line:
  `refresh MODE: total= unchanged= updated= lapsed= not-on-chain= failed= non-canonical=`.
- Never removes anything. Default `-refresh-interval` stays 1s.

### 8. Client-visible state

- No proto change in 0.7.2 (`nameserviceproto` lives in any-sync). `IsNameAvailable` already returns
  `NameExpires` for a taken name: clients can show "expired" when `NameExpires < now`.
- Follow-up (separate ticket, any-sync): add `nameExpires` (and optionally a state enum) to
  `NameByAddressResponse` so reverse lookups can show it too.

## Code to remove / simplify

`confirmNotRegistered`, finalized-block confirmation for removals, tombstone writes,
`PurgeTombstones` + CLI, expiry/lapse/daily scheduling (`expiryCheckAt`, `expiredRepairInterval`,
lapse branch of `needsRefresh`), `retryLater`/not-final bookkeeping if nothing uses it any more.
`contracts.FinalizedBlock` may stay in the contracts package if other code uses it; otherwise leave
it, unused, to keep the diff focused.

## Tests

- Run against the local replica set: `ANY_NS_TEST_MONGO="mongodb://localhost:27018/?replicaSet=rs0"
  go test -p 1 ./...` (container `any-ns-test-rs` is running). Also `go build ./...`, `go vet ./...`,
  `make run-linter` if available. Known pre-existing local failures:
  `contracts TestNormalizeEnsip1`.
- Replace the tombstone/finality tests with tests for:
  - a lapse keeps owner/AnyID/space, sets `lapsed`, updates `name_expires` and the block, no
    enrichment reads, idempotent; an older lapse observation does not overwrite a newer record.
  - not registered + stored record -> kept, no backoff loop; no stored record -> `ErrNameNotRegistered`.
  - a removed record is served as taken by `IsNameAvailable` and blocks registration.
  - lookups of expired/lapsed/removed records do not queue refreshes; `RefreshNeeded` ones do.
  - `repairOnce` skips (and fixes `repair_at` of) records due only under the old rules, without
    contract calls.
  - exponential backoff grows and resets.
  - `AdminNameRegisterSigned`: other owner / empty owner / removed record -> rejected; same owner
    -> allowed; no record -> allowed; cache error -> rejected.
  - `-restore-tombstones`: restores owners from an Extended JSON file, lists missing, dry run writes
    nothing, a record that is no longer removed is not touched.
  - `-release-name`: dry run prints, apply deletes exactly that record.
  - GO-7482 tests (`anynsaarpc/getoperation_cache_test.go`) keep passing.
- README "Name cache" section rewritten to this behaviour (rollout + rollback below).

## Rollout

1. Deploy 0.7.2 on both ns nodes (puppet `pkg::any-ns-node: 0.7.2`). From now on no tombstones,
   removed records are served as taken, no lookup/time-driven chain reads.
2. `anynsnode -c <config> -restore-tombstones -restore-from ns-cache-20261007.jsonl` (dry run), then
   with `-refresh-apply`. Expect `tombstones=~160 restored=~160 missing=0` (the prod count grows
   until 0.7.2 is deployed). The export must be the one taken before 0.7.1 was deployed.
3. `anynsnode -c <config> -refresh-cache -rpc-url <alchemy> -refresh-interval 2s` (dry run), then
   with `-refresh-apply`: fixes the stale renewals (~650 on prod), marks lapsed names.

Rollback to 0.6.9: just deploy it. 0.6.9 serves any record as taken and ignores the new fields.

## Follow-ups (not in 0.7.2)

- A strict reservation check on the user-operation path (`CreateUserOperation`): the check in
  `GetDataNameRegister*` is advisory only (client-supplied owner, opaque calldata).
- Redacting provider URLs (API keys) inside go-ethereum transport errors.
- A worker throttle for completed operations of names that are not cached (`GetOperation` path, no
  lease).
- An expected-owner flag for `-release-name`.
- A backoff for stale reads of a lagging backend (an older block).
- Carrying the poll time in `refreshRequest` for the coalescing of re-reads.

- any-sync proto: `nameExpires` / state in `NameByAddressResponse`.
- A provider fallback (Alchemy) for the live nodes' chain reads.
- Names with a leading `_` cannot be normalized (STD3); decide ENSIP-15 or exclusion.
- pp-node logs an empty `version` at startup.
