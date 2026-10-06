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

The Mongo `cache` collection mirrors the contracts; with `readFromCache: true` the lookups are served
from it. The rule: a name is reported as available, and a cached record is dropped, only when the
chain confirmed it. When unsure, the cached record is served as taken and refreshed in the background.

- **Reads.** Every read and the cache key use the canonical spelling of the name (the
  normalization of the registration, `ensip15validation`). A registry owner without a registrar
  expiry is a failure (retried), never a lapse. A refresh reads the latest block header once (number, hash and timestamp as the node
  reports them) and pins every contract read to that block by its hash. A name is registered if the
  registry has an owner and `nameExpires + 90 days (grace) >= block.timestamp`.
- **Writes.** One record per name (the unique index on `{name: 1}`; the node accepts an existing one
  whatever its name, e.g. `name_1`, creates it if it is missing, refuses to start if it is not
  unique). Every write is one Mongo transaction (majority, primary, bounded) that replaces the
  record only with a newer observation: the higher block wins. The node refuses to start without a
  replica set unless `allowUnsafeStandalone` is set (local development only).
- **Removal.** A name becomes available only when the latest block says it is not registered *and*
  a re-read at the finalized (fallback: safe) block confirms it. Only the background refresh and the
  backfill read the finalized block; a request (`readFromCache: false`, `GetOperation`) keeps the
  record and hands the name to the background. The record then becomes a
  tombstone at the finalized block: lookups treat it as a cache miss, reverse lookups skip it, and
  an older observation can not replace it. Otherwise the record stays as it is.
- **Lookups** never read the contracts and never write. A record that can be stale (expired,
  lapsed, incomplete, an old tombstone) is served as it is and handed to a background worker through
  a non-blocking in-memory queue (dropped when full).
- **Background refresh.** The worker refreshes under a lease in Mongo shared by the ns nodes; a
  failure backs off. Every record that needs a refresh is due in Mongo (`repair_at`, indexed):
  incomplete records, registrations when they expire, expired ones when they lapse and once a day,
  and scheduled re-reads. A periodic scan (`repairIntervalSec`) takes a bounded batch, the longest
  waiting first.
- **GetOperation** answers as before (GO-7482). A completed operation's name is refreshed at the
  latest block: in the poll if it is not cached yet (as before), in the background if it is
  (the poll never waits for that). Re-reads are scheduled at +5 and +30 minutes (lagging providers,
  reorgs past the Sepolia finality). If the background refresh fails, the record is marked
  `refresh_needed`; its data stays.

```
cache:
  // the cache needs a replica set (transactions). the node refuses to start on a standalone Mongo
  // unless this is set: local development ONLY, concurrent writes are not ordered there
  allowUnsafeStandalone: false

  // seconds between the periodic scans of the background refresh. 0: 60, negative: off
  repairIntervalSec: 0
```

Maintenance (one-off runs of the node binary; all are dry runs unless `-refresh-apply` is set):

- `-dedupe-cache`: verifies the unique index on `{name: 1}`. An existing unique one is accepted
  (`unique name index=exists`). A missing one is created (with `-refresh-apply`). A
  non-unique one, or duplicates of a name, is an error (exit 1); nothing is ever dropped or deleted.
- `-refresh-cache`: re-reads every cached name (tombstones too) with the same decisions as the node.
  `-refresh-interval` is the delay between two names (default 1s). Exits 1 if any name failed,
  including names whose finalized block could not be read. `not-final` (the finalized block does not
  confirm a removal yet) is counted separately and is not a failure: run it again later.
  `non-canonical` counts records cached under a spelling other than the canonical one (e.g.
  `Foo.any`): the canonical name is refreshed, such a record is left as it is (taken) for an
  operator to look at.
- `-purge-tombstones`: deletes the tombstones before a rollback (see below). It refuses (exit 1) while
  the cache has incomplete records without an owner; it lists them.

Rollout (both ns nodes share the cache):

1. Mongo must be a replica set (prod: `rs0`).
2. `anynsnode -c <config> -dedupe-cache` (prod: expect `unique name index=exists`).
3. Upgrade **both** ns nodes together. In the short mixed window an old node still serves a lapsed
   name as taken (today's behaviour) and still writes records without the new fields.
4. Only after both run the new version: `anynsnode -c <config> -refresh-cache` (dry run), then with
   `-refresh-apply`. It removes lapsed names (confirmed at a finalized block), fixes stale records
   (including what an old node wrote in the mixed window) and schedules every record for the
   periodic scan. Re-run until `failed=0`; `not-final` names are picked up by the background.

Rollback to a version older than GO-7567. Old versions read ANY record of a name as a taken name, a
tombstone too, and never repair a record without an owner:

1. `anynsnode -c <config> -purge-tombstones` (dry run), then with `-refresh-apply`. If it lists
   incomplete records, run `-refresh-cache -refresh-apply` first.
2. Deploy the old binaries on both nodes.
3. The new nodes can write tombstones until they stop: run step 1 once more (with the new binary, a
   one-off process) after the last new node stopped.

Old versions ignore the new fields and the `repair_at` index (they can be dropped later); the unique
index on `{name: 1}` stays (old versions upsert by name).

Not in this version (follow-ups): pinning the post-operation read to the operation receipt's block
(and the archive-state dependency that brings), a per-operation "cache synced" flag, verifying a
changed record against the finalized state (replaced here by the scheduled re-reads), removal
watermarks / a delete mode instead of tombstones, lookup-triggered Mongo writes, a CI job with
Mongo fail points.

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
