# Anytype Naming Service node

This global singleton node provides access to AnyNS smart contracts. You can call smart contracts directly or by using _this_ dRPC service. 

## Building and Running
1. To build: `make build`
2. To run: `go run ./cmd --c=config-sepolia.yaml`

## .yml Config file
Please see example in the '/etc' subfolder.

### Contracts section

```
contracts:
  // use your own geth node or Infura/Alchemy/Moralis/etc API
  geth_url: https://sepolia.infura.io/v3/XXX

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/ENSRegistry.json
  registry: 0xc0D3c96aE923Da6b45E6d4c21a0424730a20BCA9

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/AnytypeResolver.json
  resolver: 0x34F9c5CB9b6dcc036e045a15af20CEdC0dE4dcB2

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/AnytypeRegistrarControllerPrivate.json
  private_controller: 0x45bA047AD44e35FbF5A1375F79ea3872ceDB1732

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/AnytypeNameWrapper.json
  name_wrapper: 0xFe69BF9B3fD69d09977b37b5953C8B43687f3B23

  // Admin address
  admin: 0x61d1eeE7FBF652482DEa98A1Df591C626bA09a60
  
  // Admin key
  admin_pk: XXX
```