# Anytype Naming Service node
This global singleton node provides access to AnyNS smart contracts. You can call smart contracts directly or by using _this_ dRPC service. 

## Building and Running
1. To build: `make build`
2. To run: `go run ./cmd --c=NODE_CONFIG`
3. To run as a client: `go run ./cmd --c=CLIENT_CONFIG --cl --cmd=COMMAND --params=PARAMS_JSON`

## Available client commands

### 1. is-name-available
Check if name is available. If not - it will return information
Parameters: `'{ "FullName": "xxx.any"}'`.
Example: `go run ./cmd --c=config-client.yaml --cl --cmd=is-name-available --params='{ "FullName": "xxx.any"}'`

### 2. name-register
Create an operation to register a new name.
Parameters: `'{ "FullName": "suppa.any", "OwnerAnyAddress": "A6WVkd1MxX1i7hGQCcDhMFvfEzokPppRzxve2wdhTZ8jZTio", "OwnerEthAddress": "0xe595e2BA3f0cE990d8037e07250c5C78ce40f8fF", "SpaceId": "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu"}'`.

## .yml Config files
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