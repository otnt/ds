# Usage

1.Install binary

``go install``

2.Run main service

``$GOPATH/bin/main [name] [local port]``

``$GOPATH/bin/main Bob 8080``

``$GOPATH/bin/main Alice 8081``

3.Send request

Post a photo. Data has format

``
{
  "ImageData" : base64 encoded string
}
``

``curl -H "Content-Type: application/json" -X POST -d '{"ImageData":"1912352"}' http://localhost:8081/post``

Fetch all photos.

``curl -X GET http://localhost:8080/fetchAllPosts``
