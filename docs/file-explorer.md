# Problems
## line numbers
theres no clear cut mechanism for disabling line numbers only on certain windows.
we disable line numbers globally by default so they dont bring noise into the grid and can be rendered independently
in the frontend. 
that also means we cannot rely on neovim window options to fetch an up to date value at runtime.
we need to somehow store that information only in go or completely rethink this line number business.

## tabs
tabs dont clean up properly when closed. they leave behind some artifacts of their grid like line numbers
or buffer content.
