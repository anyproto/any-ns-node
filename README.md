# Any Naming System node
Please see [Any Naming System repository](https://github.com/anyproto/any-ns) for rationale and more info.

This global singleton node provides access to AnyNS smart contracts. You can call smart contracts either directly or by using _this_ dRPC service. 

## Building and Running
1. To build: `make deps build`
2. To run: `go run ./cmd --c=NODE_CONFIG`
3. To run as a client: `go run ./cmd --c=CLIENT_CONFIG --cl --cmd=COMMAND --params=PARAMS_JSON`

## Available client commands

### 1. is-name-available
Check if name is available. If not - it will return information
Parameters: `'{ "FullName": "xxx.any"}'`.
Example: `go run ./cmd --c=config-client.yml --cl --cmd=is-name-available --params='{ "FullName": "xxx.any"}'`

### 2. name-register
Create an operation to register a new name.
Parameters: `'{ "FullName": "suppa.any", "OwnerAnyAddress": "A6WVkd1MxX1i7hGQCcDhMFvfEzokPppRzxve2wdhTZ8jZTio", "OwnerEthAddress": "0xe595e2BA3f0cE990d8037e07250c5C78ce40f8fF", "SpaceId": "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu"}'`.

## .yml config files
Please see example in the 'etc' subfolder.
NOTICE: in order to call methods as an Admin - `account.signKey` should be used to sign messages.
That's why you can run N ns nodes with different `account.peerId` and `account.peerKey` but with SAME `account.signKey`.

### Contracts section

```
contracts:
  // use your own geth node or Infura/Alchemy/Moralis/etc API
  gethUrl: https://sepolia.infura.io/v3/XXX

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/ENSRegistry.json
  ensRegistry: 0xc0D3c96aE923Da6b45E6d4c21a0424730a20BCA9

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/AnytypeResolver.json
  resolver: 0x34F9c5CB9b6dcc036e045a15af20CEdC0dE4dcB2

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/AnytypeRegistrarImplementation.json
  registrarController: 0x6BA138bb7B1Bdea2B127D55D7C8F0DC9467b424E

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/AnytypeRegistrarControllerPrivate.json
  registrarControllerPrivate: 0x45bA047AD44e35FbF5A1375F79ea3872ceDB1732

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/AnytypeNameWrapper.json
  nameWrapper: 0xFe69BF9B3fD69d09977b37b5953C8B43687f3B23

  // Admin address
  admin: 0x61d1eeE7FBF652482DEa98A1Df591C626bA09a60
  
  // Admin key
  adminPk: XXX
```

### Name cache (`cache` section)

The Mongo `cache` collection is the record of who holds a name; with `readFromCache: true` the
lookups are served from it. Registrations only happen through Anytype (the registrar controller is
private), so every on-chain change of a name is the result of one of our operations. The rules:

- **A name stays reserved to the identity that registered it, forever**, even when it has expired
  or lapsed on chain (expired + 90 days grace). A cached name is always reported as taken (with its
  `NameExpires`: clients can show "expired" when it is in the past). Moving a name to another
  identity is a support action (`-release-name`).
- **Nothing is ever removed or wiped automatically.** Only an operator deletes a record.
- **Chain reads are driven by operations, not by lookups or time.** The chain is read to learn the
  result of our own operations, never to decide that a name is free. Expiry is a comparison of
  `name_expires` with the clock, it needs no chain read.

Details:

- **Reads.** Every read and the cache key use the canonical spelling of the name (the
  normalization of the registration, `ensip15validation`). A refresh reads the latest block header
  once and pins every contract read to that block by its hash. The registrar decides: registered
  (the grace period too; also without a registry owner, e.g. reclaimed to address(0)), lapsed, or
  never registered (no registry owner and no expiry). A registry owner without a registrar expiry
  is a failure (retried).
- **Writes.** One record per name (the unique index on `{name: 1}`; the node accepts an existing one
  whatever its name, e.g. `name_1`, creates it if it is missing, refuses to start if it is not
  unique). Every write is one Mongo transaction (majority, primary, bounded) that changes the record
  only with a newer observation: the higher block wins. The node refuses to start without a replica
  set unless `allowUnsafeStandalone` is set (local development only).
- **A lapse keeps the record.** A lapsed name is stored on its record in place (`lapsed: true`, the
  expiry, the registry owner and the block); the owner fields, AnyID, space ID and `canon` stay, and
  the owner, AnyID and space ID are not read (after a lapse the NameWrapper no longer reports the
  owner). A lapsed name that the cache does not have is not written. A later registration of the
  name (the same owner, or after a support transfer) replaces the record as usual.
- **"Never registered" never removes.** If the chain says a cached name was never registered (right
  after a registration: a lagging provider), the record stays as it is. Its due re-reads stay due
  and it backs off like a failed refresh (read again after 1, 2, 4... minutes, never every lease);
  after 6 such reads in a row (past the +30 min re-read) the due re-reads are done and the record is
  left to an operator. `GetOperation` reports `PendingOrNotFound` for a name that is neither on
  chain nor cached (GO-7482).
- **Removed records** (`removed: true`) are tombstones written by 0.7.1 (their owner fields were
  wiped). They are served as taken (without an owner), block registrations, and reverse lookups skip
  them. 0.7.2 never writes them; `-restore-tombstones` brings their owners back.
- **Registrations.** `AdminNameRegisterSigned` looks the canonical name up in the cache first (any
  record: lapsed, removed, under another spelling). A record of another identity, or one without an
  owner, rejects the request (`the name is reserved for another identity`); the same identity is
  allowed (a re-registration after a lapse, the payment node's renew-past-grace fallback). A cache
  that can not be read rejects it too (the payment node retries). Renewals are not checked.
  On the user-operation path the check is **advisory only**: `GetDataNameRegister*` refuse to build
  the calldata of a client-signed registration for a reserved name, but the owner they check is the
  one the client sends, and `CreateUserOperation` (opaque signed calldata) is not checked at all.
- **Changes** (a new registration, another owner, a renewal) are read again at +5 and +30
  minutes, whoever wrote them (not by `-refresh-cache`: the live nodes would run them on their
  provider). A read never wipes the owner: a read whose enrichment failed (or a name reserved
  without a registry owner, whose owner can not be read) keeps the owner fields, AnyID and space ID
  of the stored record that it could not read (the EOA owner of a wallet only for the same wallet);
  an incomplete record stays marked until a complete read. It never replaces a complete record of
  the same block. Every record stores its canonical spelling (`canon`).
- **Lookups** never read the contracts and never write. Only a record that needs a refresh
  (`refresh_needed`: an incomplete one, or one whose refresh after an operation failed) is handed to
  a background worker through a non-blocking in-memory queue (dropped when full). Expired, lapsed
  and removed records are never handed to it.
- **Background refresh.** The worker refreshes under a lease in Mongo shared by the ns nodes. Due in
  Mongo (`repair_at`, indexed) are only `refresh_needed` records and the scheduled re-reads; a
  periodic scan (`repairIntervalSec`) takes a bounded batch, the longest waiting first. A record
  that 0.7.1 scheduled under its rules (at its expiry, its lapse, daily) is rescheduled by the scan
  from its fields, without a chain read; such records do not count toward the batch (at most 500
  per round). A failed refresh backs off exponentially per record (`refresh_failures`: 1 min,
  doubled with every failure in a row, at most 24 h), reset by a successful write of the record. A
  replacing write holds the record for a minute (`refresh_next_at`).
- **GetOperation** answers as before (GO-7482). A completed operation's name is refreshed at the
  latest block: in the poll if it is not cached yet, in the background if it is (the poll waits
  neither for the contracts nor for a write: the re-reads are stored on the record in the
  background, bounded by 2 seconds, coalesced and capped). Re-reads are scheduled at +5 and +30
  minutes (lagging providers, reorgs past the Sepolia finality). If the background refresh fails,
  the record is marked `refresh_needed`; its data stays. A record held by another refresh still gets
  the operation's re-reads (the worker stores them). Repeated polls within a minute of a refresh do
  not read the chain again.

```
cache:
  // the cache needs a replica set (transactions). the node refuses to start on a standalone Mongo
  // unless this is set: local development ONLY, concurrent writes are not ordered there
  allowUnsafeStandalone: false

  // seconds between the periodic scans of the background refresh. 0: 60, negative: off
  repairIntervalSec: 0
```

Maintenance (one-off runs of the node binary; all are dry runs unless `-refresh-apply` is set; they
run in this order: `-dedupe-cache`, `-release-name`, `-restore-tombstones`, `-refresh-cache`):

- `-dedupe-cache`: verifies the unique index on `{name: 1}`. An existing unique one is accepted
  (`unique name index=exists`). A missing one is created (with `-refresh-apply`). A
  non-unique one, or duplicates of a name, is an error (exit 1); the index is never dropped.
  It then migrates records cached under a non-canonical spelling (e.g. `Foo.any`, written by an
  old node), one transaction each: no canonical record → renamed to the canonical name (data
  kept, marked for a refresh); a live, complete canonical record at least as new → the alias is
  deleted (only if the canonical record has its chain block: a legacy record without one has
  unknown freshness); otherwise both stay, the canonical one is marked for a refresh, and they are
  listed (`kept:`) for an operator or a later run. Until then the node warns at start, an alias
  keeps its name taken (every lookup without a canonical record, or with a removed one, checks for
  one, through a case- and accent-insensitive index `name_ci` the node creates; the candidates are
  checked by the normalization), reverse lookups prefer the canonical record, and the background
  refreshes the canonical name instead of the alias.
  `-dedupe-cache -refresh-apply` also gives every record its `canon`: the alias check finds an
  alias by it (e.g. a punycode spelling, which the collation can not equate).
  Every other alias stays and keeps its name taken: `-dedupe-cache` lists them (`kept:`),
  `-refresh-cache` counts them (`non-canonical=N`), the node warns at start. An operator resolves
  them by hand (`db.cache.deleteOne({name: "<alias spelling>"})`) if the canonical record is right.
- `-release-name <name>`: the support transfer of a name. Normalizes the name and prints its record
  (`owner_any_address`, `owner_eth_address`, `owner_scw_eth_address`, `space_id`, `name_expires`,
  `lapsed`, `removed`); with `-refresh-apply` deletes exactly that record (one `DeleteOne` by `_id`,
  `observed_at` and `owner_any_address` as printed, logged at warn with the full record; if the record
  changed in between nothing is deleted: `the record changed, run again`). Records of the name under
  other spellings are listed, never deleted (they keep the name taken: delete them by hand). Exits 1
  if the name is not cached. After a release the name can be registered for a new identity.
- `-restore-tombstones -restore-from <file>`: gives the 0.7.1 tombstones their owners back from a
  `mongoexport` of the cache (JSON lines, MongoDB Extended JSON, e.g. the nightly dump; not
  `--jsonArray`). For every removed record (one transaction each, only while it is still removed):
  found in the export with an `owner_any_address` → the owner fields and `space_id` from the export,
  `name_expires` the later of the export's and the tombstone's (a restore with `name_expires` 0 is
  logged), `lapsed` if it has lapsed by now (otherwise `refresh_needed`: the background reads it
  once), `canon` if missing, `removed` cleared; the block of the tombstone stays. Not found, or
  without an owner → `  missing: <name>`; records of the name (or of a spelling of it) with
  different owners in the export → `  conflict: <name>`; both are left as they are (taken, without
  an owner). Live records whose owner differs from the export's are listed as
  `  owner-differs: <name>` (a name taken by another identity while 0.7.1 had freed it?), never
  written: for support. Output: `restore DRY RUN|APPLIED: tombstones=N restored=N skipped=N
  missing=N conflicts=N owner-differs=N` (tombstones = restored + skipped + missing + conflicts;
  `skipped`: no longer removed at the write); exits 1 if any are missing, conflicting or differ.
- `-refresh-cache`: re-reads every cached name with the same decisions as the node, and never
  removes anything. `-refresh-interval` is the delay between two names (default 1s). Output:
  `refresh MODE: total= unchanged= updated= lapsed= not-on-chain= failed= non-canonical= unnormalizable=`.
  `lapsed` counts the records marked (or kept) lapsed, `not-on-chain` the cached names the chain
  says were never registered (kept, for an operator to look at), `non-canonical` the records cached
  under another spelling (the canonical name is refreshed, such a record is left as it is),
  `unnormalizable` the names the normalization rejects (expected: e.g. names with a leading `_`;
  not read, left as they are, not a failure). A changed record gets no re-reads from this run.
  Exits 1 if any name failed: run it again until `failed=0`.
- `-rpc-url <url>` (maintenance runs only): the contracts provider for this process instead of
  `contracts.gethUrl`, so that a full `-refresh-cache` does not spend the live nodes' provider quota.
  At startup its `eth_chainId` must match `accountAbstraction.chainID` of the config (or, without
  one, the chain id of `contracts.gethUrl`), otherwise the run refuses to start. Only its host is
  logged (go-ethereum transport errors can still carry the full URL: see the follow-ups).

Rollout of 0.7.2 (both ns nodes share the cache):

Before every `-refresh-apply` run take a `mongoexport` (JSON lines) or `mongodump` of `ns.cache`.

1. Deploy 0.7.2 on **both** ns nodes. From then on no tombstones are written, removed records are
   served as taken, and there are no lookup- or time-driven chain reads. Do the next steps only after
   no 0.7.1 node runs any more (a 0.7.1 node keeps writing tombstones).
2. **Right away** (until then the owners of tombstoned names are rejected when they register their
   own name again): `anynsnode -c <config> -restore-tombstones -restore-from ns-cache-<date>.jsonl`
   (dry run), then with `-refresh-apply`. The export must be JSON lines (`mongoexport` without
   `--jsonArray`). Expect `missing=0 conflicts=0`; a missing or conflicting name stays taken without
   an owner (a registration of it is rejected) until an operator decides (`-release-name`, or
   restore it by hand). Review the `owner-differs:` list with support (names that another identity
   may have registered while 0.7.1 had freed them).
3. `anynsnode -c <config> -refresh-cache -rpc-url <another provider> -refresh-interval 2s` (dry
   run), then with `-refresh-apply`: fixes the stale renewals, marks the lapsed names. Some
   `unnormalizable` names are expected. Re-run until `failed=0`.

Rollback to 0.6.9: just deploy it on both nodes. 0.6.9 serves any record as taken (a removed one too)
and ignores the new fields (`lapsed`, `refresh_failures`) and the indexes, so the data stays safe;
the server-side reservation check of registrations goes away with it. Never roll back to 0.7.1: it
frees lapsed names and wipes their owners again.

Not in this version (follow-ups): `nameExpires` (and a state) in `NameByAddressResponse` (any-sync
proto), a strict reservation check on the user-operation path (`CreateUserOperation`), redacting
provider URLs inside go-ethereum transport errors, a provider fallback for the live nodes' chain
reads, pinning the post-operation read to the operation receipt's block, names with a leading `_`
(STD3), a CI job with Mongo fail points.

Tests of the cache need a Mongo: a standalone one on `localhost:27017` (the replica set tests skip),
or a one-node replica set. The fail point tests (stalled Mongo commands) need `enableTestCommands`:

```
docker run -d --name any-ns-test-rs -p 27018:27018 mongo:7 --replSet rs0 --port 27018 --bind_ip_all \
  --setParameter enableTestCommands=1
mongosh --port 27018 --eval 'rs.initiate({_id:"rs0",members:[{_id:0,host:"localhost:27018"}]})'
ANY_NS_TEST_MONGO="mongodb://localhost:27018/?replicaSet=rs0" go test -p 1 ./cache/ ./anynsrpc/ ./anynsaarpc/
```

## Contribution

 Thank you for your desire to develop Anytype together!

 ❤️ This project and everyone involved in it is governed by the [Code of Conduct](https://github.com/anyproto/.github/blob/main/docs/CODE_OF_CONDUCT.md).

 🧑‍💻 Check out our [contributing guide](https://github.com/anyproto/.github/blob/main/docs/CONTRIBUTING.md) to learn about asking questions, creating issues, or submitting pull requests.

 🫢 For security findings, please email [security@anytype.io](mailto:security@anytype.io) and refer to our [security guide](https://github.com/anyproto/.github/blob/main/docs/SECURITY.md) for more information.

 🤝 Follow us on [Github](https://github.com/anyproto) and join the [Contributors Community](https://github.com/orgs/anyproto/discussions).

---

Made by Any — a Swiss association 🇨🇭

Licensed under [MIT](./LICENSE.md).
