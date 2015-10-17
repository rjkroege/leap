The `search_test.go` unit test requires the `test_index` file to exist
and be valid. This should be checked in.  It's built with the `cindex`
tool which is prerequisite to having `leap` do anything useful anyway.

When adding new tests that require additional content, add the
content to the `test_data` tree and build the index for the first timelike so:

	CSEARCHINDEX=`{pwd}^/test_index cindex test_data

Rebuild:
	
	CSEARCHINDEX=`{pwd}^/test_index cindex 



