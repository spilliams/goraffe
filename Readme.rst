*******
Goraffe
*******

graphing with go. get it?

.. image:: doc/goraffe.png

Description
===========

Very much a work in progress. I'm using this to learn about go tools, brush off
old graphviz skillz, and perhaps make something useful to other Go developers.

This codebase does not currently use go modules, and is meant to work only
inside a ``$GOPATH``, and run against other repos in the ``$GOPATH``.

Installation
============

.. code-block:: console

   $ go get -u github.com/spilliams/goraffe/goraffe

Usage
=====

.. note::

   This CLI command is built with `cobra <https://github.com/spf13/cobra/>`__,
   so all of its subcommands have a ``-h|--help`` option for displaying
   documentation, as well as a ``-v|--verbose`` option for printing more output
   (to ``stderr``).

Imports
-------

``goraffe imports`` is a command that builds a tree of package names, connected
by how they import each other. The various options and flags of imports are
described in its help function (``goraffe imports --help``), but here are some
example use cases:

.. code-block:: console

   $ goraffe imports <parent directory> <root package> [<root package> ...] [flags]

The basic command with ``<parent directory>`` and ``<root packages>`` will
build the whole tree of everything inside the parent directory, starting from
the named roots.

.. code-block:: console

   $ goraffe imports <parent> <root> --keep <other> [--keep <other> ...] --grow 2

The "keep/grow" flags will let you zero in on a specific package in the tree,
and see importers and importees N levels away (in this example, 2).

.. code-block:: console

   $ goraffe imports <parent> <root> --branch <other> [--branch <other> ...]

The "branch" flag will let you track all the import paths between the root(s)
and the named package(s).

Referrers
---------

.. code-block:: console

   $ goraffe referrers <package>.<func>

TODO
====

1. any kind of tests
2. optionally add a legend to the graphviz output?
3. try again on go modules (first attempt was super slow to import everything)
4. include externals should get dialed back: main concern is the first level of
   external imports. I don't care about the second level.

   This could manifest in the two separate calls to ``shouldInclude``, one from
   the main add func, one from the imports loop.
5. think more about applicability in a shared repo setting: product-server,
   product-shared, product-agent are 3 repos that all import each other, and
   while I don't care about deep externals, I do care about deep imports of
   other repos. Maybe that's just a manipulation of the parent directory and
   root packages though.
6. ``referrers``: stderr and stdout should get captured in buffers, then we
   can parse them and do the next step: graphviz
7. Gosh golly, why doesn't ``guru callers`` like me?

call tracer
-----------

``goraffe calltree <pkg> <func>`` can we introspect on function calls within
the context of a package?

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

update: can i use something like a Language Server Protocol for getting this
info? https://langserver.org/

the working scripts I'm using for this dev:

.. code::

   guru referrers goraffe/cmd/root.go:#865
   guru -scope github.com/spilliams/goraffe/... callees goraffe/cmd/root.go:#865
   guru -scope github.com/spilliams/goraffe/...,-github.com/spilliams/goraffe/vendor callers goraffe/cmd/root.go:#865

   goraffe -v referrers goraffe/cmd.initLogger

Resources to Explore
--------------------

- davecheney's `glyph <https://github.com/davecheney/junk/tree/master/glyph>`__

- `gddo-server <https://github.com/golang/gddo/blob/master/gddo-server/graph.go>`__
- https://github.com/kisielk/godepgraph

- https://groups.google.com/forum/#!forum/gonum-dev
- https://www.gonum.org/post/introtogonum/
- `gonum...dot <https://github.com/gonum/gonum/tree/master/graph/encoding/dot>`__

- https://github.com/sourcegraph/go-langserver or https://github.com/golang/go/wiki/gopls ?

- `VSCode Go extension <https://github.com/microsoft/vscode-go>`__
- `guru doc <https://docs.google.com/document/d/1_Y9xCEMj5S-7rv2ooHpZNH15JgRT5iM742gJkw5LtmQ/edit>`__
- `guru source <https://github.com/golang/tools/blob/master/cmd/guru/callers.go>`__
