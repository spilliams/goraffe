*******
Goraffe
*******

graphing with go. get it?

.. image:: doc/goraffe.png

Description
===========

Very much a work in progress. I'm using this to learn about go tools, brush off
old graphviz skillz, and perhaps make something useful to other Go developers.

TODO
====

1. any kind of tests
2. add a legend to the graphviz output
3. rewrite the command so that "repo" is the first arg, then "roots" (do I need
   "keeps" to be separate from "roots"?). Allow for ellipse nodes that
   correspond to folders instead of packages.

call tracer
-----------

``goraffe calltree <pkg> <func>`` can we introspect on function calls within the context of a package?

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
