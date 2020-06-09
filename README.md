# Slicer API

![Go](https://github.com/SlicerChat/API/workflows/Go/badge.svg) ![Go Report Card](https://goreportcard.com/badge/github.com/SlicerChat/API)

## Database

Currently, Slicer is being rewritten to use MongoDB.
This is due to:

- Slow speeds with Cassandra;
- Having to use raw CQL queries;
- General usability improvements.

Cassandra *is* a good choice, but its benefits aren't,
well, beneficial enough to make using it over MongoDB
viable.

## Documentation Status

Currently, the API is subject to many breaking changes.
Documentation is being held off until a more stable
release.
