# radixmmap: sort massive files by prefix in memory

In the example we will sort a file in chronological order, where first 19 bytes of each line are assumed to be RFC3339 date.

Example

```.bash
> cat /tmp/test
2011-01-10T15:30:45Z,bla
2009-01-10T15:30:47Z,abc
2009-01-10T15:30:45Z,def

> ./radixmmap -s /tmp/test
2009-01-10T15:30:45Z,def
2009-01-10T15:30:47Z,abc
2011-01-10T15:30:45Z,bla
```

# How?

First we load file into memory & then use [radix sort][radix] to sort by first 19 bytes of each line.

# Why?

The idea is to sort big files as fast as possible with as little overhead as possible.

I find that [memory mapped files][mmap] allow for optimal loading of the file:
this way OS allocates just as much memory as needed.

Prior to this utility I have been using [sort][sort] command found in shells:


```.bash
LC_ALL=C sort --parallel=16 -t, -k1 -S100% /tmp/test
```

but found it quite memory hungry.

[mmap]: https://en.wikipedia.org/wiki/Memory-mapped_file
[radix]: https://en.wikipedia.org/wiki/Radix_sort
[sort]: https://en.wikipedia.org/wiki/Sort_(Unix)
