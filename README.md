# radixmmap: sort massive files by prefix in memory

In this example we will sort a file in chronological order, where first 19 bytes of each line are assumed to be RFC3339 date.

Example 1:

```.bash
> cat /tmp/test
2011-01-10T15:30:45Z,bla
2009-01-10T15:30:47Z,abc
2009-01-10T15:30:45Z,def

> ./radixmmap /tmp/test
2009-01-10T15:30:45Z,def
2009-01-10T15:30:47Z,abc
2011-01-10T15:30:45Z,bla
```

Example 2, multiple csv files with header:

```.bash
cat /tmp/a /tmp/b
date,string
2009-01-10T15:30:45Z,def
2009-01-10T15:30:47Z,abc
2011-01-10T15:30:45Z,bla
date,string
2020-01-10T15:30:45Z,key
2005-01-10T15:30:47Z,lkj
1999-01-10T15:30:45Z,zxc

> ./radixmmap -skip-header /tmp/a /tmp/b
1999-01-10T15:30:45Z,zxc
2005-01-10T15:30:47Z,lkj
2009-01-10T15:30:45Z,def
2009-01-10T15:30:47Z,abc
2011-01-10T15:30:45Z,bla
2020-01-10T15:30:45Z,key
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

# Main idea behind the implementation

This implementation uses as little RAM as possible without compromising on performance too much.

Additionally to loading file in RAM, we need 16 bytes per line to remember
where the lines start and where they end (8 bytes to remember start
position and 8 bytes to remember end position). For a file with 1.2 billion
lines this results in 19.2 GB overhead.

# Benchmark

Currently sorts 44GB file using 63.2GB RAM, 16 cores in 19 minutes 37 seconds:


```.bash
> go build && time ./radixmmap -d sorted.csv bigfile.csv
real    19m37.992s
user    96m27.955s
sys     1m7.665s
```

In comparison `sort` takes 29 minutes and 51 seconds, and uses >80GB of RAM:

```.bash
> time LC_ALL=C sort --parallel=16 -t, -k1 -S100% -o sorted.csv bigfile.csv
real    29m51.346s
user    71m34.478s
sys     3m17.947s
```

The file has ~1.22 billion lines.

# Credits

Big thanks to [edsrzf][edsrzf] and [twotwotwo][twotwotwo] for providing
underlying implementations for mmap & radix sort respectively.

[mmap]: https://en.wikipedia.org/wiki/Memory-mapped_file
[radix]: https://en.wikipedia.org/wiki/Radix_sort
[sort]: https://en.wikipedia.org/wiki/Sort_(Unix)
[edsrzf]: https://github.com/edsrzf/mmap-go
[twotwotwo]: https://github.com/twotwotwo/sorts
