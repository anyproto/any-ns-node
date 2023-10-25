# Any Naming System node
Please see [Any Naming System repository](https://github.com/anyproto/any-ns) for rationale and more info.

This global singleton node provides access to AnyNS smart contracts. You can call smart contracts either directly or by using _this_ dRPC service. 

## Building and Running
1. Prereq: `go install go.uber.org/mock/mockgen@latest`
2. To build: `make build`
3. To run: `go run ./cmd --c=NODE_CONFIG`
4. To run as a client: `go run ./cmd --c=CLIENT_CONFIG --cl --cmd=COMMAND --params=PARAMS_JSON`

## Available client commands

### 1. is-name-available
Check if name is available. If not - it will return information
Parameters: `'{ "FullName": "xxx.any"}'`.
Example: `go run ./cmd --c=config-client.yaml --cl --cmd=is-name-available --params='{ "FullName": "xxx.any"}'`

### 2. name-register
Create an operation to register a new name.
Parameters: `'{ "FullName": "suppa.any", "OwnerAnyAddress": "A6WVkd1MxX1i7hGQCcDhMFvfEzokPppRzxve2wdhTZ8jZTio", "OwnerEthAddress": "0xe595e2BA3f0cE990d8037e07250c5C78ce40f8fF", "SpaceId": "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu"}'`.

## .yml Config files
Please see example in the 'etc' subfolder.

### Contracts section

```
contracts:
  // use your own geth node or Infura/Alchemy/Moralis/etc API
  gethUrl: https://sepolia.infura.io/v3/XXX

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/ENSRegistry.json
  registry: 0xc0D3c96aE923Da6b45E6d4c21a0424730a20BCA9

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/AnytypeResolver.json
  resolver: 0x34F9c5CB9b6dcc036e045a15af20CEdC0dE4dcB2

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/AnytypeRegistrarImplementation.json
  registrar: 0x6BA138bb7B1Bdea2B127D55D7C8F0DC9467b424E

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/AnytypeRegistrarControllerPrivate.json
  privateController: 0x45bA047AD44e35FbF5A1375F79ea3872ceDB1732

  // https://github.com/anyproto/any-ns/blob/master/deployments/sepolia/AnytypeNameWrapper.json
  nameWrapper: 0xFe69BF9B3fD69d09977b37b5953C8B43687f3B23

  // Admin address
  admin: 0x61d1eeE7FBF652482DEa98A1Df591C626bA09a60
  
  // Admin key
  adminPk: XXX
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
