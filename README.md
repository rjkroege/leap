Leap
====

What is it? A program for leaping around with Acme via Alfred. The
idea: I press the cmd-space bar, get the Alfred bar, type the key word
(`leap` say) and then a string. The string will find a buffer in Acme
and search through the buffer. The string will look like any string
that can be right-moused to search. The notion is to more rapidly
navigate files in a context where I don't have a mouse.

The basic scheme here could be extended to search the code as well.
Note though that this is explicitly not in scope for phase 1. Phase 2
will add code searching locally. Phase 3 will add client-server code
searching.

How does it work?
===

*  Alfred workflow launches leap.
*  leap reads command line
*  parses command line
*  talks to Acme and gets auto-completion suggestions
*  generates response list

NB: This will be stateless unless required to be otherwise.

Syntax
===
Type any sequence that would be an acceptable right mouse button action in Acme but
get instant feedback of possibilities.


Examples
===
Usage examples. Write this

		leap Foo

Searches for a file whose name (or path) contains `Foo`. 

		leap Foo:/blah/

Seaches forward in a file whose name (or path) contains `Foo` for the string `blah` 
and selects it. 

		leap cf:/blah/

Searches forward in a `content/foo.c` for `blah` so long as `content/foo.c` is the only
file uniquely identified by `cf`. 



Tasks
====

*  <strike>Setup a trial workflow. I can reuse the one from the author of the class.</strike>
* <strike> Insert the code from the example into the trial workflow</strike>
*  <strike>Explore the contents of the example workflow</strike>
*  <strike>Get the example running from command line</strike>. Testing is accomplished by having a copy of the plist file in the directory.
*  <strike>Get the example running from Alfred</strike>
*  <strike>Read the Acme index file from the command line</strike>
* <strke>Parse Acme index file</strike>
* Propose acme index file reading as change to rsc
*  <strike>Replace the list of entries with the acme index files</strike>


Issues
====
I'm perplexed: what happens when we actually hit enter... What does
that program get? In particular, I need to do stuff to Acme then. And
that means that I need to know which option the user actually chose.
Right? Some of my existing workflows (pinboard search for example)
resolve this.

I'm going to need another command. It's not clear what it gets as
arguments. But something from the first command. And I wish that
Alfred supported styling the strings somehow.

Aside: maybe it does. You have no idea what they do with the strings
that I return. I can try various things. Like returning an HTML string
and see what happens.

Indeed. There are two different commands. The picker component and the
doer. Let's accept that there will be two different commands. Which
means what exactly?

I note in passing that the search examples will warp the cursor and
enlarge the window if I make the *opener* program use the `plumb`
interface. If and only if we are using a file. No. See `addr`,
`dot=addr` and `show`. It can be coerced easily.

Future
===
I have read about [commad-T](https://wincent.com/products/command-t).
This is where I want to go. And how I want to search for files.
Imagine the following:

*  I have a *current project* which is the root of the tree being considered for
searching.
*  I type a string of letters. It can be broken down into substrings. We return a 
list of matching entities. Open windows can be folded into this with a different
icon.
* The algorithm would appear to be a bit subtle. Which is cool. Yay.
Given a string, each substring could match differently. And then we
need to merge the matches together and prune the result. I need to
refine this further. Sounds like an inverted index to me.

Approach
----

		Split the paths up at the `/`
		Group path names at each level by common prefixes. Stop grouping once every
		prefix with 2 or more paths has been identified.

		empty the candidate list

		for each prefix sub-string of the search string
			add path component matching to the candidate list
			recurse with search suffix on paths rooted at each path component

		stop as soon as candidate list has more than 20 entries
			
Thoughts
----
The above doesn't seem quite right. This is subtle!

Work an example. Assume some root `.`

		a

And I get:

*   all `./a*`
*   all  non-root leaf files matching `a.*`

If I type

		ab

*  `./ab*`
*  all non-root files matching `ab*`
*  `./a*/b*`

If I type

		abc

*  `./abc*`
*   all non-root files matching `abc*`
*  `./a*/bc*`
*  `./a*/b*/c*
*   `./ab*/c*

I might want to be more clever with file name matching. 
Let's watch the peepcode video. Maybe not. Doesn't seem to exist now.

If I type
	
		abcd

*  `./abcd*`
*  non-root files matching `abcd*` (Or something more clever)
*  set . to `./a` and recurse on `bcd`
*  set . to `./ab` and recurse on `cd`
*  set . to `./abc` and recurse on `d`

How do I sort these together?

I should order alphabetically. I should bound the number included in display with
an ellipsis if the count of that block exceeds a threshold. I can start with just a list?

Observation: a regular expression works pretty well here. No particular fanciness needed.
I can be more clever like I suppose if I really want.

Follow On Tasks
===

Phase 1
-----

*  <strike>extract the list of filenames</strike>
*  <strike>elide the start of the names so that they fit better</strike>
*  <strike>convert the typed stuff into regexp </strike>
*  <strike>apply the regexp to the list of filenames</strike>
*  <strike>don't add an entry twice if matched multiple times by different regexps</strike>
*  add icons to the result
*  package this up in some kind of rational way: we need the file opening
*  figure out the additional keys *what does this mean?* This means: understand how to use the other fields in the XML that gets shipped to Alfred.
*  support auto-complete | enter doing something different.


Phase 2
----

*  take better advantage of `/` characters to improve matching (will need to refine this)
*  re-write this document pending upstreaming
*  upstream this content
*  refactor the code to be nicer: there is a pipeline here
	*  Get acme index data
	*  Get file matches data (I want an interface for fetching)
	*  Search acme index data and create intermediate type for matches
	*  Search file data and create intermediate type for matches
	*  Merge all intermediate type entries together, sort and de-duplicate
	*  Append additional properties to intermediate type
	*  Generate Alfred output from intermediate type

Please remember that phase 2 (file matching) is not part of this exercise. I need
to finish phase 1 first.

Please resort the above based on what's in phase 1 or phase 2. And minimize the
work imposed.





