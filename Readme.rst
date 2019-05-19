*******
Goraffe
*******

graphing with go. get it?

Description
===========

Very much a work in progress. I'm using this to learn about go tools, brush off
old graphviz skillz, and perhaps make something useful to other Go developers.

TODO
====

- ``goraffe imports`` add a ``--single <this>`` flag to get just the things that import this, and the things this imports
- ``goraffe imports`` add an ``--outline <...>`` flag to outline certain nodes (such as "nodes only imported by one other package")
- ``goraffe calltree <pkg> <func>`` can we introspect on function calls within the context of a package?

.. code::

   ./scripts/build.sh && goraffe -v imports github.com/spilliams/goraffe/cli --prefix github.com/spilliams/goraffe | dot -Tsvg > graph.svg && open graph.svg
   ./scripts/build.sh && goraffe -v imports github.com/spilliams/goraffe/cli --prefix github.com/spilliams/goraffe --single github.com/spilliams/goraffe/cli/cmd/imports | dot -Tsvg > graph.svg && open graph.svg
   
call tracer
-----------

I have one or more funcs I wish to trace
I want to trace them back to roots (so the graph doesn't get so out of whack)
Thankfully, the API naturally terminates at the top-level packages
``endpoints`` etc.
For callers in a test func, I can check the filename to see the ``_test.go``
and terminate those traces at the ``test`` root?

So for something like ``log.Debugf``, which is called in the CLI and services
and other "top-level" packages, I'd like to be able to specify which roots I
care about.
Sort of a "trace the func up to the roots, then pick some of the roots to keep,
and trickle down a 'keep' flag, then trim away all the stuff that isn't 'keep'"

Methodology
-----------

I very much like the simplicity of ``dot``, but sometimes it gets...hard to
read. We'll likely want a product that hosts an HTTP service, with an API and a
front-end. The front-end will have a bunch of d3 stuff on it.

Resources to Explore
--------------------

- read that book on dataviz
- `gddo-server <https://github.com/golang/gddo/blob/master/gddo-server/graph.go>`__
- `davecheney/graphpkg <https://github.com/davecheney/graphpkg>`__
- davecheney's `glyph <https://github.com/davecheney/junk/tree/master/glyph>`__
- `gonum...dot <https://github.com/gonum/gonum/tree/master/graph/encoding/dot>`__, or `awalterschulze/gographviz <https://github.com/awalterschulze/gographviz>`__
- https://codefreezr.github.io/awesome-graphviz/#libs-for-go
- https://groups.google.com/forum/#!forum/gonum-dev
- https://www.gonum.org/post/introtogonum/
- `runtime Caller <https://golang.org/pkg/runtime/#Caller>`__
