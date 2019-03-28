# Ripple
**Ripple** is a cli client for a decentralised IRC-like system.

## Usage
```
usage: ripple [<flags>] <command> [<args> ...]

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).
  --tui   Run the application in terminal UI mode

Commands:
  help [<command>...]
    Show help.

  start <addr>
    Start a new swarm

  connect <listen-addr> <connect-addr>
    Connect to an existing swarm
```
**Ripple** provides 2 main commands:
- *start* which starts a brand new swarm
- *connect* which connects to an existing swarm

Once connected to the swarm, we need an address on which to listen for new
connections, which is what the first argument is for

After connecting to a swarm, we can send messages by typing in the terminal.

We can change our nickname for other peers by entering `!nick newname` in the terminal.

## Terminal UI
Ripple also comes with a terminal UI, which can be used by passing the `--tui` flag.