How should search work?

On init we can read in all files and add them to the index if they haven't been added already.
We can also check their modified times and if they are newer re-read them and update the index
The can all be done through a goroutinue. If files are bigger than a certain threshold we can just
skip them since opening them and storing them in the index would be very resource heavy.

We can provide the search interface through an api route and then use javascript on the frontend to call it and filter the files available.

We can keep a table of the last time a file was scraped in memory and have a goroutine that checks if a new version of that file has been written

- We can probably combine the bleve search file and the mdserver general storage file, this should help move the in-memory index last scraped table to be persistent

* Add search result match highlighting - we could preview 4-5 lines of whatever matched
