// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package params

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main Pistchain network.
var MainnetBootnodes = []string{
	"enode://f23a0d53c638d2b3ccc1c017413a50f09fa04b066c9840c644f6db8b015de6503f1bd6a8067e5c4a6847ad9c1a44d55333500ffd91e04d0e06cff9aebbdd7674@3.34.205.183:30111", // CN
	"enode://ff93f032ead7bbfb2593efb010ecab206c2f5cfc1f2e976b6ae6923ccfdcb99811493a27ff7a527b97e3c1592970db3bfff4317e2df69922b11b0f82cdd677f3@3.35.71.3:30111",    // US WEST
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
	"enode://f23a0d53c638d2b3ccc1c017413a50f09fa04b066c9840c644f6db8b015de6503f1bd6a8067e5c4a6847ad9c1a44d55333500ffd91e04d0e06cff9aebbdd7674@3.36.129.45:30111",
	"enode://d4bff15e626b8819d9f1cafde4ad0da99d4203a4ce447692fb0560ce6ef2c6a9fb2b0f70d11c77a769952724c12799d44b6dcd260659583bafad106ab9266561@15.165.36.87:30111",
	"enode://7f9b9d1d920fa4a098090835bf806746c1341edf4dc8c8bdb70457be56dd799491631048d03f8a7fca5df14b4886c5772e8ca7cfcfaca63f5ecc96d1c01d2047@15.165.34.198:30111",
	"enode://fbb0bed7934eea4e802fcdce22debca68ddbd80e6e3b046cc2c0c54fdc6932ce8dd66a688de36a80899952cf070a992346170f8645f01cd74d5b6c88a3ea08e8@3.35.77.52:30111",
}

// DevnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the dev Pistchain network.
var DevnetBootnodes = []string{
	"enode://f23a0d53c638d2b3ccc1c017413a50f09fa04b066c9840c644f6db8b015de6503f1bd6a8067e5c4a6847ad9c1a44d55333500ffd91e04d0e06cff9aebbdd7674@127.0.0.1:30111",
	"enode://d4bff15e626b8819d9f1cafde4ad0da99d4203a4ce447692fb0560ce6ef2c6a9fb2b0f70d11c77a769952724c12799d44b6dcd260659583bafad106ab9266561@127.0.0.1:30111",
	"enode://7f9b9d1d920fa4a098090835bf806746c1341edf4dc8c8bdb70457be56dd799491631048d03f8a7fca5df14b4886c5772e8ca7cfcfaca63f5ecc96d1c01d2047@127.0.0.1:30111",
	"enode://fbb0bed7934eea4e802fcdce22debca68ddbd80e6e3b046cc2c0c54fdc6932ce8dd66a688de36a80899952cf070a992346170f8645f01cd74d5b6c88a3ea08e8@127.0.0.1:30111",
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{
	"enode://ebb007b1efeea668d888157df36cf8fe49aa3f6fd63a0a67c45e4745dc081feea031f49de87fa8524ca29343a21a249d5f656e6daeda55cbe5800d973b75e061@39.98.171.41:30315",
	"enode://b5062c25dc78f8d2a8a216cebd23658f170a8f6595df16a63adfabbbc76b81b849569145a2629a65fe50bfd034e38821880f93697648991ba786021cb65fb2ec@39.98.43.179:30312",
}
