# Quill

Scheduler of operations on in-memory data. The rabbit hole has gone to far.

Meant for parallelizing expensive operations on a collection of data by taking advantage of explicit read/write decelerations.  

Imagine your nasty data as a traditional database. You need to make some queries on it. You need a bunch of different operations to run on this database. The database is actually a struct though. In memory. How do you design your system to not block access to your precious ugly struct as much as possible?