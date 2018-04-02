fuzz: proto-fuzz.zip
	go-fuzz -bin=./ninep-fuzz.zip -workdir=./fuzz-data

proto-fuzz.zip:
	go-fuzz-build github.com/azmodb/ninep/proto
