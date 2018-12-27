module github.com/sdeoras/rpi-automation

require (
	cloud.google.com/go v0.34.0
	github.com/golang/protobuf v1.2.0
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/tensorflow/tensorflow v1.12.0
	go.opencensus.io v0.18.0 // indirect
	golang.org/x/net v0.0.0-20181213202711-891ebc4b82d6
	google.golang.org/api v0.0.0-20181217000635-41dc4b66e69d // indirect
	google.golang.org/grpc v1.17.0
)

replace github.com/tensorflow/tensorflow => ../tensorflow
