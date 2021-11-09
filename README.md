## pist chain

pist chain is a truly fast, permissionless, secure and scalable public blockchain platform. [more](https://github.com/pistchain/pist/)

<a href="https://github.com/pistchain/pist/blob/main/COPYING"><img src="https://img.shields.io/badge/license-GPL%20%20pist-lightgrey.svg"></a>

## Building the source


Building pist requires both a Go (version 1.14 or later) and a C compiler.
You can install them using your favourite package manager.
Once the dependencies are installed, run


    make gpist

or, to build the full suite of utilities:

    make all

## Running pist

Going `pist -h` can get help infos.

### Running on the pist main network

```
$ pist console
```


### Running on the pist Chain singlenode(private) network

To start a g instance for single node,  run it with these flags:

```
$ pist --singlenode  console
```

This command will:

* Start gpist with network ID `189`
* Start up gpist's built-in interactive console,
  (via the trailing `console` subcommand) through which you can invoke all official [`web3` methods]



### Running on the PIST Chain test network

To test your contracts, you can join the test network with your node.

```
$ gpist --testnet console
```

The `console` subcommand has the exact same meaning as above and they are equally useful on the testnet too. Please see above for their explanations if you've skipped here.

Specifying the `--testnet` flag, however, will reconfigure your Geth instance a bit:

* Test network uses different network ID `188`
* Instead of connecting the main PIST chain network, the client will connect to the test network, which uses testnet P2P bootnodes,  and genesis states.

### Configuration

As an alternative to passing the numerous flags to the `gpist` binary, you can also pass a configuration file via:

```
$ gpist --config /path/to/your_config.toml
```

To get an idea how the file should look like you can use the `dumpconfig` subcommand to export your existing configuration:

```
$ gpist --your-favourite-flags dumpconfig
```


### Running on the PIST Chain singlenode(private) network

To start a g instance for single node,  run it with these flags:

```
$ gpist --singlenode  console
```

Specifying the `--singlenode` flag, however, will reconfigure your Geth instance a bit:

* singlenode network uses different network ID `400`
* Instead of connecting the main or test pistchain network, the client has no peers, and generate fast block without committee.

 
