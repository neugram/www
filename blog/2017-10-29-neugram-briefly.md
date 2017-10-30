# Neugram, briefly

_October 2017_

If you program a lot in Go, you may find Neugram interesting.

I started working on Neugram because as I spent more time programming
in Go I found myself writing a larger fraction of the scripts better
suited to Python or Perl in Go.
My daily programming came to be dominated by bash and Go.

The problem is, bash is an awkward language for a ~100 line program,
and sometimes so is Go.
While it is thoroughly enjoyable to use the same standard library in
scripts as in big complex projects, Go is slower to work in than
Python and Perl for a few reasons:

- No read-eval-print-loop.
- No [shebang](https://en.wikipedia.org/wiki/Shebang_(Unix)) support.
- Lots of unhelpful error handling.
(Go's explicit error handling is wonderful in large programs, but
in small scripts where all you write is
`if err != nil { log.Fatal(err) }` it is a drag.)

At first glance these look like features missing from Go.
So I went about figuring out how to add them.
Turns out, they are missing quite deliberately.

## REPL

The grammar of Go needs to be changed to support line-by-line
evaluation.
Top-level constructions in a `.go` file are declarations, not
statements.
There is no sequence in the declarations, all are evaluated
simultaneously across all the files in a package.
A declaraction on an earlier line in a file can happily refer to a
name declared later in the file.
If you want to type Go declarations into a REPL, nothing can execute
until you declare the package done.

So the first thing you need to do to define a REPL for Go is to step
down a level.
Instead of declaractions, process statements.
Pretend everything typed into the REPL is happening inside the
`func main() {}` of a Go program.
Now there is a sequence of events and statements can be evaluated as
they are read.

This shrinks the set of programs you can write dramatically.
In Go there is no way to define a method on a type inside a function
(that is, using statements).
There is a good reason for this: all the methods of a type need to be
defined simultaneously, so that the method set of a type doesn't
change over time.
It would lead to a whole new class of confusing errors if you could
write:

```go
func main() {
	type S string
	var s S
	_, ok1 := s.(io.Reader)
	func (S) Read(b []byte) (int, error) { ... }
	_, ok2 := s.(io.Reader)
	fmt.Println(ok1, ok2) // Prints: false, true
}
```

That is why you cannot write that in Go.

So for the language to be REPL-compatible it needs serious grammar
surgery, which would make a REPL possible, but hurt the readability
of big complex programs.

Neugram has its own [its own statement-based method syntax](https://github.com/neugram/ng/blob/master/eval/testdata/method2.ng),
which diverges in a small but significant way from Go. (Though it
won't be properly functional until the
[Go generating backend](https://github.com/neugram/ng/issues/5)
is complete.)

## Error handling

Explicit vs. implicit error handling is a contentious issue, but it
is a safe bet that if you have chosen to use Go you strongly favor
explicit error handling. The language does not make it easy to avoid
handling your errors.

Unfortunately, there is one place where even a strong supporter of
explicit error handling can admit the process is tedious: when
writing "all or nothing" scripts.
That is, programs that either follow the narrow success path
completley, or if they step off the path even slightly exit
immediately in error.
Small Python scripts follow this process by default, and bash script
authors often do by placing `set -e` at the top of their scripts.

Indeed, a common source of consternation for Java or Python
programmers coming to Go is discovering that the small program they
attempted to write to try out Go ended up needlessly wordy.
This:

```python
#!/usr/bin/python
f = open("hello.txt","w" 
f.write("Hello, World!") 
f.write("Next line.")
f.close()
```

Becomes:

```go
package main

import (
	"os"
	"log"
	"fmt"
)

func main() {
	f, err := os.Create("hello.txt")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := fmt.Fprintf(f, "Hello, World!\n"); err != nil {
		log.Fatal(err)
	}
	if _, err := fmt.Fprintf(f, "Next line.\n"); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
```

This is not quick to write and is one of the few places where
explicit error handling in Go is unhelpful.
Our script is no more robust for handling these errors explicitly.

Neugram is designed to help with scripts like these.

Neugram is `set -e` for Go:

```neugram
#!/usr/bin/ng

import "os"
import "fmt"

f := os.Create("hello.txt")
fmt.Fprintf(f, "Hello, World!\n")
fmt.Fprintf(f, "Next line.\n")
f.Close()
```

The elided errors in the Neugram program will turn into panics if
they are non-nil. The result is a script which is not too much
wordier than the python version, while taking advantage of all my
Go knowledge. If you are a Go programmer, this may be interesting
to you.

## What's next

Go is syntactically a big language. Neugram's front end has to match
it, and that's a lot of work, much of which is still to do. The
*mountain of bugs* is my first priority.

If you want to help out, try writing something and file all the bugs
you run into (you will) on the [issue tracker](https://github.com/neugram/ng/issues).

After that there are several other language extensions I'm interested
in, focusing on making Neugram a good language for data analysis.
In particular:

- [Operator overloading](https://github.com/neugram/ng/issues/2),
  it would be a terrible idea in Go, but I think it's a good fit
  for Neugram.
- [Table (matrix) types](https://github.com/neugram/ng/issues/1)
- [A generic implicit type parameter, _num_](https://github.com/neugram/ng/issues/3)
- [Go generating backend](https://github.com/neugram/ng/issues/5)

I am hesitant to enter any discussion about programming languages,
especially given how much more work Neugram needs to be generally
usable. But there is only so long I can work quietly on a project,
so I may as well at least admit it exists. Let me know what you
think on the
[mailing list](https://groups.google.com/forum/#!forum/neugram).

_By David Crawshaw ([@davidcrawshaw](https://twitter.com/davidcrawshaw))_
