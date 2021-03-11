# Guppy

Guppy is a compression algorithm designed to reduce the size of N-body simulations. For a typical cosmological simulation, guppy can reduce snapshots by a factor ten to fifteen in exchange for small, tightly controled errors on the positions and velocities of particles.

This github repository contains a command line program which can create compressed `.gup` files. It also contains libraries for reading these files in C, Python 3, and Go in the `c/`, `python/`, and `go/` folders.

Instructions for installing and using guppy can be found in the `docs/` folder:

* [docs/install.md](https://github.com/phil-mansfield/guppy/blob/main/docs/install.md): Installing the guppy command line program.
* [docs/run.md](https://github.com/phil-mansfield/guppy/blob/main/docs/run.md): Running the guppy command line program.
* [docs/c.md](https://github.com/phil-mansfield/guppy/blob/main/docs/c.md): Importing and using the C library for reading `.gup` files.
* [docs/python.md](https://github.com/phil-mansfield/guppy/blob/main/docs/python.md): Importing and using the Python library for reading `.gup` files.
* [docs/go.md](https://github.com/phil-mansfield/guppy/blob/main/docs/go.md): Importing and using the Go library for readign `.gup` files
.
Guppy is currently in verison 0.0.1. It may experience breaking changes and may have major bugs. The scientific paper on guppy is still being written.

## FAQs

**Where should I go for help if I can't compile guppy or if I think it isn't working correctly?**

Please look through the Issues tab and see if someone else has encountered this problem before. If not, feel free to submit an issue there.

**How should I acknowledge guppy if I use `.gup` files as part of a scientific paper?**

We would appreciate a citation to the guppy code paper, XXXXXXXX, somewhere in your methods section.

**Can I copy part of the code from the guppy repository into my software project?**

Yes, guppy uses a permissive MIT license, so you can copy any part of it into your project, even if you plan to publish or sell it. However, you will need to include the text of the MIT license in that part of the code.

**What should I do if my langauge doesn't have a `.gup` reader?**

You will need to write it yourself. You can use the existing readers and the guppy code paper as references and should feel free to contact me if you have questions. If you write a reader, contact me and I will link to it here.
