# Coding Guidelines

* Keep all lines to 80 or fewer characters if tabs are rendered with four spaces.
* The exception to this is large error messages, so users can always `grep` or them.
* Comment all functions.
* Add a brief note at the top of each file explaining what's in that file unless that file contains the package comment.
* Test all functions except for the truly trivial ones.
* All errors in `lib/` subpackages should return potentially user-viewed error messages rather than calling `panic()` or using the `error` subpackage.
* `lib/` funcitons should use `ExternalError` if the error could potentially be caused by malformed user input. `InternalError` should be used if it's an issue of internal inconsistency that requires a code dive.
* All `ExternalError` calls should not require a code dive to understand.
