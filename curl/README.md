# ECH-enabled `curl`

This is a custom build of `curl` with ECH support from the [DEfO project](https://github.com/defo-project). It is useful for manually testing ECH functionality against specific web servers.

## Prerequisites (for building on macOS)

* Homebrew
* `automake`
* `libtool`
* `pkg-config`
* `libpsl`

## Building

A helper script, `build-curl.sh`, is provided to automate the build process for `curl` and its dependency, `openssl`.

To build the ECH-enabled `curl`, run the script from the project root and provide an output path:

```sh
./curl/build-curl.sh <output_directory>
```

For example, to build `curl` and place the output in the `workspace` directory:

```sh
./curl/build-curl.sh workspace
```

The script will download the source code for `openssl` and `curl`, build them, and install the final binaries in the specified output directory.

For more details on how to use `curl` with ECH, see the [official documentation](https://github.com/defo-project/curl/blob/master/docs/ECH.md).

## Verifying the build

To test that your custom `curl` build is working correctly, run it against the DEfO test server:

```sh
./workspace/output/bin/curl" --ech=true --doh-url https://1.1.1.1/dns-query 'https://test.defo.ie/echstat.php?format=json' | jq
```

Example output:
```json
{
  "SSL_ECH_OUTER_SNI": "public.test.defo.ie",
  "SSL_ECH_INNER_SNI": "test.defo.ie",
  "SSL_ECH_STATUS": "success",
  "date": "2025-11-06T19:36:47+00:00",
  "config": "min-ng.test.defo.ie"
}
```

You may need to specify the `LD_LIBRARY_PATH` on Linux:

```sh
LD_LIBRARY_PATH="$(pwd)/workspace/lib" ./workspace/bin/curl --version
```
