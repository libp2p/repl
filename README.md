# A libp2p REPL (*)

***>>> This is a fun, hacky toy project (for now). 🎈 <<<***

(*) [REPL](https://en.wikipedia.org/wiki/Read%E2%80%93eval%E2%80%93print_loop) = Read-Eval-Print-Loop, or interactive shell.

[![](https://img.shields.io/badge/made%20by-Protocol%20Labs-blue.svg?style=flat-square)](https://protocol.ai)
[![](https://img.shields.io/badge/project-libp2p-yellow.svg?style=flat-square)](https://libp2p.io/)
[![](https://img.shields.io/badge/freenode-%23libp2p-yellow.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23libp2p)
[![GoDoc](https://godoc.org/github.com/libp2p/repl?status.svg)](https://godoc.org/github.com/libp2p/repl)
[![Coverage Status](https://coveralls.io/repos/github/libp2p/repl/badge.svg?branch=master)](https://coveralls.io/github/libp2p/repl?branch=master)
[![Build Status](https://travis-ci.com/libp2p/repl.svg?branch=master)](https://travis-ci.com/libp2p/repl)
[![Discourse posts](https://img.shields.io/discourse/https/discuss.libp2p.io/posts.svg)](https://discuss.libp2p.io)

> A small libp2p REPL that starts an embedded host and offers an interactive
> menu to trigger actions.
> 
> We use it in libp2p workshops and demos to accompany
> technical walkthroughs. We might add a shell mode, and extend it in a lot of
> magical ways if if it takes off.

## Instructions

1. Clone this repository.
2. cd into this directory.
3. `go build .`
4. Run the binary with environment variable `LIBP2P_ALLOW_WEAK_RSA_KEYS=true`.

## License

Dual-licensed under MIT and ASLv2, by way of the [Permissive License Stack](https://protocol.ai/blog/announcing-the-permissive-license-stack/).