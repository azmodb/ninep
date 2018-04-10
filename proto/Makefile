fuzz: proto-fuzz.zip
	go-fuzz -bin=./proto-fuzz.zip -workdir=./fuzz-data

proto-fuzz.zip:
	go-fuzz-build github.com/azmodb/ninep/proto

clean:
	rm -rf fuzz-data proto-fuzz.zip
