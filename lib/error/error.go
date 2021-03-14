/*package error contains simple funcitons for reporting Guppy errors.
*/
package error

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
)

// External reports an error to stderr and kills the function. It should be used
// when an error is something a user could reasonbly be expected to fix through
// changes in configuration/data/environement. It has the same signature at the
// standard fmt.*printf() functions.
func External(format string, a ...interface{}) {
	log.Printf("Guppy exited early with the following error:\n" + format, a...)
	os.Exit(1)
}

// Internal reports an error to stdout along with a strack trace and  kills the
// function. It should be used when the error requires a code dive to fix. It
// has the same signature at the standard fmt.*printf() functions.
func Internal(format string, a ...interface{}) {
	log.Println("Guppy exited early with the following error:")
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintf(os.Stderr, "\n\n")
	debug.PrintStack()
	os.Exit(1)
}
