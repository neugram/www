# Neugram Design Principles

_3 November 2017_

Neugram is a new project with a long way to go.
To help stay focused over a long project and to make it accessible
to potential new contributors, this post documents some
design principles for Neugram.

There are lots of considerations in designing a language, more
than it would make sense to list in a post specific to one language.
Instead, this post focuses on the unusual aspects of Neugram,
and the way those aspects constrain future design.

### _1. Go statements should be equivalent Neugram statements._

Neugram is for that space of programs you would like to write in Go
but the language doesn't fit, and no other language quit fits either:
tasks that are too complex for bash, are possible in Python or Perl
or Ruby but would benefit from static types or other Go features
or packages.

One of the major reasons to use Neugram is to take advantage of all
that Go knowledge some of us keep in our heads for writing larger
programs, so Neugram should be as unsurprising as possible to a Go
programmer.
To that end, as Neugram gets features that aren't in Go, the features
should be extensions to the language, not incompatible changes.

### _2. If a feature makes sense as a Go proposal, propose it for Go._

Neugram is for scripting, and it is deeply intertwined with Go.
You shouldn't find yourself replacing a Go program with a Neugram
program.
If Go was already the right tool for the job, then it still is.
Features that would make Go better for a job belong in a Go proposal,
not Neugram.

This is a very useful principle for evaluating features as it makes
it easy to reject some big proposals independent of their intrinsic
value.
For example, generics (specifically some kind of parametric
polymorphism) would be really useful.
But generics would also be really useful in Go and are a major part
of the discussion around Go 2.
So it is inappropriate to consider generics for Neugram.
If you have something in mind for generics, write a proposal for Go 2!

### _3. The higher-level the language, the more surprises._

Neugram is a higher-level language than Go, and as such is allowed to
be surprising in ways we wouldn't expect in Go.

For example, if you see

```
	x[y]
```

in Go code, with no other context you know you are indexing into an
array, slice, or map named `x` with an existing value `y`.
As such, the amortized cost of the expression `x[y]` is O(1).
This is a wonderful feature of Go when reading through unfamiliar
code where performance matters, and one I would never want to lose.

However in a higher-level language, the value of knowing the
performance of square brackets matters less to me than the flexibility
of being able to index into a custom type backed by something
interesting. I mind far less in my script if `x[y]` triggers an RPC
or runs an SQLite statement, in fact that sounds like a useful and
interesting feature.

So operator overloading, a feature I would never want to see in Go,
is very much on the table for Neugram.

### _4. Minimize novelty to maximize casual programming._

I do not write significant scripts every day. Not even every week.

The rest of the time, my programming is inside existing large
programs, in languages like Go or C++.
When it comes time to write a script, I can pick from a number
of languages I have used before, like bash, Perl, or python.
In each of these languages I have written tens of thousands of lines,
but I do so infrequently enough that by the time it comes to write my
next script I have forgotten the libraries and syntax to do common
tasks.
How do I find and replace in strings, or format dates and times, or
manipulate files? My programming time ends up divided between the
text editor writing the program and a web browser searching for
_"how do I _ in {Perl,python,bash}"_. This is frustrating.

The goal of Neugram is that if you are programming in Go daily,
you can take all your knowledge of Go syntax and its standard
libraries and use them when doing the casual programming of
writing a script.

To this end, Neugram should minimize its innovation. If there is
a syntax trick, it should aim to be One Big Trick (like the shell's
`$$`) that is easy to remember rather than a lot of unusual cases.

_By David Crawshaw ([@davidcrawshaw](https://twitter.com/davidcrawshaw))_
