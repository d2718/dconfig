# dconfig
A Go package for reading unix configuration files of the `OPTION=value` format.

I use this package in many of my programs.

I think the best instruction I can give for how to use the package is just
to look at the `example/example.go` (and its accompanying
`example/example.conf`), for an example, and then take a peek at the source
if you're looking for specifics. I realize this sounds like a cop-out, but
it's a simple, straightforward package.

### Update 2017-04-22

This update changed the method of operation and the format of the function
calls to be more like the standard library `flag` package. This change allows
appropriately-written packages to avoid stepping on each other's feet despite
having identically-named options. Unfortunately, this change is
backward-incompatible, and breaks all programs written prior to it.

