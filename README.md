# NINEP

[![GoDoc](https://godoc.org/github.com/azmodb/ninep?status.svg)](https://godoc.org/github.com/azmodb/ninep)
[![Build Status](https://travis-ci.org/azmodb/ninep.svg?branch=master)](https://travis-ci.org/azmodb/ninep)

Package ninep serves network filesystems using the 9P2000.L protocol. The package provides types and routines for implementing 9P2000.L servers and clients.

> **WARNING:** This software is new, experimental, and under heavy
> development. The documentation is lacking, if any. There are almost
> no tests. The APIs and source code layout can change in any moment.
> Do not trust it. Use it at your own risk.A
>
> **You have been warned**


## References

- 9P2000.L protocol [overview](https://github.com/chaos/diod/blob/master/protocol.md)
- [VirtFS](https://landley.net/kdocs/ols/2010/ols2010-pages-109-120.pdf) -- A virtualization aware File System pass-through, Jujjuri et al, 2010.
- Plan 9 from Bell Labs - [Section 5](https://9p.io/sys/man/5/INDEX.html) - Plan 9 File Protocol, 9P, Plan 9 Manual, 4nd edition, 2002.
- [v9fs](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/plain/Documentation/filesystems/9p.txt): Plan 9 Resource Sharing for Linux
- Plan 9 Remote Resource Protocol Unix Extension [experimental-draft-9P2000-unix-extension](http://ericvh.github.io/9p-rfc/rfc9p2000.u.html), Van Hensbergen, 2009.


## Authors

See list of [CONTRIBUTORS](https://github.com/azmodb/ninep/blob/master/CONTRIBUTORS) who participated in this project.


## License

This project is licensed under the ISC license - see the [LICENSE](https://github.com/azmodb/ninep/blob/master/LICENSE) file for details.
