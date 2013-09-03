Leap
====

What is it? A program for leaping around with Acme via Alfred. The idea: I press the cmd-space bar,
get the Alfred bar, type the key word (`leap` say) and then a string. The string will find a buffer in 
Acme and search through the buffer. The string will look like any string that can be right-moused
to search. The notion is to more rapidly navigate files in a context where I don't have a mouse.

The basic scheme here could be extended to search the code as well. Note though that this is
explicitly not in scope for phase 1. Phase 2 will add code searching locally. Phase 3 will add
client-server code searching.

How does it work?
===

*  Alfred workflow launches leap.
*  leap reads command line
*  parses command line
*  talks to Acme and gets auto-completion suggestions
*  generates response list

NB: This will be stateless unless required to be otherwise.

