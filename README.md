# Go License

Print the combined licenses of your imports and program.

Example:
```sh
go get github.com/howeyc/golicense

cd /path/to/go/program/source

golicense
```

This is useful to comply with the second clause of the BSD license:

* Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.

Also useful to see if a dependency has a license you don't recognize.

Basically this will go through the GOPATH in search of LICENSE* files and then
recursively go through all the imports used in the go package to find all
required licenses. It then prints a combined license file to Standard Output.

GOROOT is also used to get the License of the Go Standard Library.
