# b2 [![Build Status](https://travis-ci.org/romantomjak/b2.svg)](https://travis-ci.org/romantomjak/b2) [![Coverage Status](https://coveralls.io/repos/github/romantomjak/b2/badge.svg?branch=master)](https://coveralls.io/github/romantomjak/b2?branch=master) [![GoDoc](https://godoc.org/github.com/romantomjak/b2?status.svg)](https://godoc.org/github.com/romantomjak/b2)

Backblaze B2 Cloud Storage Command Line Client

---

## Status

This project is in development phase. I highly recommend you to use b2 with go modules.

## Install

```sh
go get -u github.com/romantomjak/b2
```

## Usage

```sh
$ export B2_KEY_ID=1234
$ export B2_KEY_SECRET=MYSECRET
$ b2 create my-globally-unique-bucket-name
Bucket "my-globally-unique-bucket-name" created with ID "123b2bucketid8"
```

## TODO

First I'll be focusing on creating Buckets. This is the progress so far:

- [x] Create a new bucket
- [ ] Delete bucket
- [ ] List all buckets
- [ ] Update settings for a bucket

## Contributing

You can contribute in many ways and not just by changing the code! If you have any ideas, just open an issue and tell me what you think.

Contributing code-wise - please fork the repository and submit a pull request.

## License

MIT
