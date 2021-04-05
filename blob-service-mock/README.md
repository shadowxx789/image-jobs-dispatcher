## Stub of "worker.blob.net" service

### DESCRIPTION
This is stub which provide API for working with blob storage

Lang version: Go 1.15<br>
Linter: GolangCI 1.35.2 with list of rules in .golangci.yaml<br>
Base docker images:

    - Building: https://github.com/theshamuel/baseimg-go-build
    - Application: https://github.com/theshamuel/baseimg-go-app


#### API v1

1. Put blob `POST: /api/v1/blob` `Headers: Content-Type: <Ther MIME content type of the blob>; Content-Length: The size of the content`
   - Request: <binary image content>
     - Ex: `curl --request POST \
     --url http://HOST:8081/api/v1/job \
     --header 'content-type: image/png' \
     --header 'content-length: 123456' \
     --data <your binary data of image>

   - Response:
        - JSON: 
          <pre>{
            "id":"1"
          }</pre>

1. Get blob data `GET: /api/v1/blob/{id}` `Headers: Content-Type: <Ther MIME content type of the blob>;Content-Length: The size of the content`
    - Request: No Body
      - Ex: `curl --request GET \
        --url http://localhost:8081/api/v1/blob/1`
   - Response:
     - Headers: `Content-Type: <Ther MIME content type of the blob>; Content-Length: The size of the content`
     - Body: `<binary image content>`
    
### Improvements for using in a real pipeline as contract/smoke tests
    1. Add cases with different MIME types
    2. Add functionality upload/get binary of images by chunks for making using network connection more balanced