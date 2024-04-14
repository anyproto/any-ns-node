package accountabstraction

const erc20ABI = `
		[
			{
				"constant": false,
				"inputs": [
					{
						"name": "_to",
						"type": "address"
					},
					{
						"name": "_amount",
						"type": "uint256"
					}
				],
				"name": "mint",
				"outputs": [],
				"payable": false,
				"stateMutability": "nonpayable",
				"type": "function"
			},
			{
				"inputs": [
					{
						"internalType": "address",
						"name": "from",
						"type": "address"
					},
					{
						"internalType": "address",
						"name": "spender",
						"type": "address"
					},
					{
						"internalType": "uint256",
						"name": "value",
						"type": "uint256"
					}
				],
				"name": "approveFor",
				"outputs": [
					{
						"internalType": "bool",
						"name": "",
						"type": "bool"
					}
				],
				"stateMutability": "nonpayable",
				"type": "function"
			}
		]	
	`

const commitABI = `
		[
			{
				"inputs": [
					{
						"internalType": "bytes32",
						"name": "commitment",
						"type": "bytes32"
					}
				],
				"name": "commit",
				"outputs": [],
				"stateMutability": "nonpayable",
				"type": "function"
			}
		]
	`

const executeABI = `
	[
		{
			"inputs": [
				{
					"internalType": "address[]",
					"name": "dest",
					"type": "address[]"
				},
				{
					"internalType": "bytes[]",
					"name": "func",
					"type": "bytes[]"
				}
			],
			"name": "executeBatch",
			"outputs": [],
			"stateMutability": "nonpayable",
			"type": "function"
		}
	]
`

const regABI = `
		[
			{
				"inputs": [
					{
						"internalType": "string",
						"name": "name",
						"type": "string"
					},
					{
						"internalType": "address",
						"name": "owner",
						"type": "address"
					},
					{
						"internalType": "uint256",
						"name": "duration",
						"type": "uint256"
					},
					{
						"internalType": "bytes32",
						"name": "secret",
						"type": "bytes32"
					},
					{
						"internalType": "address",
						"name": "resolver",
						"type": "address"
					},
					{
						"internalType": "bytes[]",
						"name": "data",
						"type": "bytes[]"
					},
					{
						"internalType": "bool",
						"name": "reverseRecord",
						"type": "bool"
					},
					{
						"internalType": "uint16",
						"name": "ownerControlledFuses",
						"type": "uint16"
					}
				],
				"name": "register",
				"outputs": [],
				"stateMutability": "nonpayable",
				"type": "function"
			}	
		]
	`

const factoryContractABI = `
		[
			{
				"inputs": [
					{
						"internalType": "address",
						"name": "owner",
						"type": "address"
					},
					{
						"internalType": "uint256",
						"name": "salt",
						"type": "uint256"
					}
				],
				"name": "getAddress",
				"outputs": [
					{
						"internalType": "address",
						"name": "",
						"type": "address"
					}
				],
				"stateMutability": "view",
				"type": "function"
			}
		]
	`

const entryPointJSON = `
		[
			{
				"inputs": [
					{
						"internalType": "address",
						"name": "sender",
						"type": "address"
					},
					{
						"internalType": "uint192",
						"name": "key",
						"type": "uint192"
					}
				],
				"name": "getNonce",
				"outputs": [
					{
						"internalType": "uint256",
						"name": "nonce",
						"type": "uint256"
					}
				],
				"stateMutability": "view",
				"type": "function"
			}
		]
		`
