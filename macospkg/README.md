This is a archive plugin for [Hugoreleaser](https://github.com/gohugoio/hugoreleaser) that runs via [execrpc](https://github.com/bep/execrpc).

**Note:** If you are running your releases from a CI Linux container, there is a remote version ofthis plugin in [macospkgremote](../macospkgremote).

There will be more and better documentation later, but 

* The main work is performed via [buildpkg](https://github.com/bep/buildpkg).
* Setting up the AWS entities needed for this, see [s3rpc](https://github.com/bep/s3rpc).