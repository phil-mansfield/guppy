# Running guppy

The guppy command line program is a configuration file-based tool which is made up of three smaller sub-programs.

* `guppy check` - Checks the contents of a configuration file and attempts to guess whether guppy will crash when executing it.
* `guppy convert` - Converts snapshot files to the `.gup` format and visa versa.
* `guppy confirm` - Confirms that a set of snapshot files matches the contents of a corresponding set of `.gup` files to within the specified error limits.

The typical pattern for a user on a large computing cluster would be:

1. Write a configuration file based on the example files in `example_configs/`
2. Log into a small interactive job session and run `guppy check` on that config file and fixing errors until the checks pass.
3. Submit a large job which runs `guppy convert`.
4. Confirm that there are no bugs in the the `.gup` files using `guppy confirm`. The truly paranoid can run a second large job to check evey particle in their snapshots, but checking a few files will usually be enough.
