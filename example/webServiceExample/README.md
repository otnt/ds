# Usage

1.Install binary

``go install``

2.Run web service

``$GOPATH/bin/webServiceExample [name] [local port]``

``$GOPATH/bin/webServiceExample Bob 8080``

``$GOPATH/bin/webServiceExample Alice 8081``

3.Send request

Post a photo. Data has format

``
{
  "ImageURL" : base64 encoded string
}
``

``curl -H "Content-Type: application/json" -X POST -d '{"ImageURL":"123"}' http://localhost:8081/addPost``

Fetch all photos.

``curl -X GET http://localhost:8080/fetchAllPosts``

Add comment.

``curl -H "Content-Type: application/json" -X POST -d '{"ImageURL":"123"}' http://localhost:8081/addComment`

Add up vote.

``curl -H "Content-Type: application/json" -X POST -d '{"ImageURL":"123"}' http://localhost:8081/upVote`

Add down vote.

``curl -H "Content-Type: application/json" -X POST -d '{"ImageURL":"123"}' http://localhost:8081/downVote`
