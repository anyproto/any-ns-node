const fs = require('fs')

const fileName = process.argv[2]
const contract = JSON.parse(fs.readFileSync(fileName, 'utf8'))
fs.writeFileSync('contract.abi', JSON.stringify(contract.abi))

let bc = JSON.stringify(contract.bytecode)
// remove " symbols from the bin
fs.writeFileSync('contract.bin', bc.replace(/[\"]/g, ''))
