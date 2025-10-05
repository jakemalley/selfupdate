module github.com/jakemalley/selfupdate

go 1.25.1

replace github.com/kr/binarydist => ./internal/binarydist

require github.com/kr/binarydist v0.1.0

require github.com/dsnet/compress v0.0.1 // indirect
