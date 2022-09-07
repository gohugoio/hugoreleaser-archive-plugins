This is a archive plugin for [Hugoreleaser](https://github.com/gohugoio/hugoreleaser) that runs via [execrpc](https://github.com/bep/execrpc) and [s3rpc](https://github.com/bep/s3rpc).

**Note:** If you are running your releases from a MacBook or similar, there is a simpler version of this plugin in [macospkg](../macospkg).

This is a server both meant to be started as a plugin from Hugorelaser, but also started as a local build server responding to build requests from e.g. a CI server:

```
go run github.com/gohugoio/hugoreleaser-archive-plugins/macospkgremote@latest localserver
```

The above will poll a AWS SQS queue for new jobs approximately once every minute. You only need to have that running when its needed, but you do have [1 million free SQS request](https://aws.amazon.com/sqs/pricing/) every month.

There will be more and better documentation later, but 

* The main work is performed via [buildpkg](https://github.com/bep/buildpkg).
* Setting up the AWS entities needed for this, see [s3rpc](https://github.com/bep/s3rpc).