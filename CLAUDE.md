# CLAUDE.md

- Comment why, not what. Only comment when it isn't obvious.

- Implementation code (`internal/*.go`) is sparsely commented — roughly one note
  per function that needs one, and most need none. Match that:
    - No doc comments on functions. There are none in the implementation; a
      function that seems to need one usually needs a better name instead
    - Put the note inside the function, directly above the line it explains
    - One to three lines. A comment that wants a paragraph is a TODO.md entry or
      a commit message
    - Terse step labels are fine in a function with phases:
      `// Resolve parent / master project`, `// Nothing to do, no volumes`

- Test code is the opposite and should stay that way: a `// TestX proves ...` doc
  comment carrying the reasoning, because a test's value is the claim it makes.
