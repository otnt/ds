# Usage

``go install``

``$GOPATH/bin/webServiceExample Bob 8080``

``$GOPATH/bin/webServiceExample Alice 8081``

``curl -H "Content-Type: application/json" -X POST -d '{"ImageData":"1912352"}' http://localhost:8081/post``

``curl -X GET http://localhost:8080/fetchAllPosts``
