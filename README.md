Hi Jin, I was bored so did it...

Func writeTar() streams the bytes for the tar into any writer. In main.go I'm writing to a gzip Writer, but you could write into your S3 uploader instead.

And I added a test, since it's pretty critical we don't mess up.

Needs review.
